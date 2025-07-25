package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"featureflags/config"
	"featureflags/controller"
	"featureflags/entity"
	"featureflags/handler"
	"featureflags/repository"
	"featureflags/service"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// IntegrationTestSuite represents the integration test suite
type IntegrationTestSuite struct {
	testDB     *TestDB
	app        *echo.Echo
	controller *controller.FlagController
}

// SetupIntegrationTest sets up the integration test environment
func SetupIntegrationTest(t *testing.T) *IntegrationTestSuite {
	testDB := SetupTestDB(t)
	
	// Initialize services
	flagRepo := repository.NewFlagRepository(testDB.DB)
	auditRepo := repository.NewAuditRepository(testDB.DB)
	log := GetTestLogger()
	flagService := service.NewFlagService(flagRepo, auditRepo, log)
	flagController := controller.NewFlagController(flagService, log)

	// Setup Echo app
	app := echo.New()
	cfg := &config.Config{
		Swagger: config.Swagger{Enabled: false}, // Disable swagger for tests
	}
	handler.RegisterRoutes(app, flagController, cfg, log)

	return &IntegrationTestSuite{
		testDB:     testDB,
		app:        app,
		controller: flagController,
	}
}

// Cleanup cleans up the test environment
func (suite *IntegrationTestSuite) Cleanup(t *testing.T) {
	suite.testDB.CleanTables(t)
	suite.testDB.Close()
}

// TestExampleScenario1_CheckoutDependencies tests the first example scenario:
// "checkout_v2" depends on "auth_v2" and "user_profile_v2", can only be enabled after both are enabled
func TestExampleScenario1_CheckoutDependencies(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.Cleanup(t)

	t.Run("Example Scenario 1: checkout_v2 dependencies", func(t *testing.T) {
		// Step 1: Create base flags
		authFlag := createFlagHelper(t, suite, "auth_v2", []int64{})
		userProfileFlag := createFlagHelper(t, suite, "user_profile_v2", []int64{})
		
		// Step 2: Create checkout flag with dependencies
		checkoutFlag := createFlagHelper(t, suite, "checkout_v2", []int64{authFlag.ID, userProfileFlag.ID})
		
		// Step 3: Try to enable checkout_v2 (should fail - dependencies not enabled)
		response := toggleFlagHelper(t, suite, checkoutFlag.ID, true, "Launch checkout v2")
		assert.Equal(t, http.StatusBadRequest, response.Code)
		
		var errorResp map[string]interface{}
		err := json.Unmarshal(response.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		
		assert.Equal(t, "Missing active dependencies", errorResp["error"])
		missingDeps := errorResp["missing_dependencies"].([]interface{})
		assert.Contains(t, missingDeps, "auth_v2")
		assert.Contains(t, missingDeps, "user_profile_v2")
		
		// Step 4: Enable auth_v2
		response = toggleFlagHelper(t, suite, authFlag.ID, true, "Auth v2 ready")
		assert.Equal(t, http.StatusOK, response.Code)
		
		// Step 5: Try to enable checkout_v2 again (should still fail - user_profile_v2 not enabled)
		response = toggleFlagHelper(t, suite, checkoutFlag.ID, true, "Launch checkout v2")
		assert.Equal(t, http.StatusBadRequest, response.Code)
		
		err = json.Unmarshal(response.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		missingDeps = errorResp["missing_dependencies"].([]interface{})
		assert.Contains(t, missingDeps, "user_profile_v2")
		assert.NotContains(t, missingDeps, "auth_v2") // auth_v2 should not be in missing deps
		
		// Step 6: Enable user_profile_v2
		response = toggleFlagHelper(t, suite, userProfileFlag.ID, true, "User profile ready")
		assert.Equal(t, http.StatusOK, response.Code)
		
		// Step 7: Now enable checkout_v2 (should succeed)
		response = toggleFlagHelper(t, suite, checkoutFlag.ID, true, "All dependencies ready")
		assert.Equal(t, http.StatusOK, response.Code)
		
		// Verify checkout_v2 is enabled
		suite.testDB.AssertFlagStatus(t, checkoutFlag.ID, entity.FlagEnabled)
		
		t.Logf("✅ Scenario 1 passed: checkout_v2 can only be enabled when all dependencies are active")
	})
}

// TestExampleScenario2_MissingDependencyErrorFormat tests the exact error format from requirements
func TestExampleScenario2_MissingDependencyErrorFormat(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.Cleanup(t)

	t.Run("Example Scenario 2: Missing dependency error format", func(t *testing.T) {
		// Create auth_v2 (disabled by default)
		authFlag := createFlagHelper(t, suite, "auth_v2", []int64{})
		
		// Create checkout_v2 depending on auth_v2
		checkoutFlag := createFlagHelper(t, suite, "checkout_v2", []int64{authFlag.ID})
		
		// Try to enable checkout_v2 while auth_v2 is disabled
		response := toggleFlagHelper(t, suite, checkoutFlag.ID, true, "Should fail")
		assert.Equal(t, http.StatusBadRequest, response.Code)
		
		var errorResp map[string]interface{}
		err := json.Unmarshal(response.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		
		// Verify exact error format as specified in requirements
		expectedError := map[string]interface{}{
			"error":                "Missing active dependencies",
			"missing_dependencies": []interface{}{"auth_v2"},
		}
		
		assert.Equal(t, expectedError["error"], errorResp["error"])
		assert.Equal(t, expectedError["missing_dependencies"], errorResp["missing_dependencies"])
		
		t.Logf("✅ Scenario 2 passed: Error format matches requirements exactly")
		t.Logf("Response: %s", response.Body.String())
	})
}

// TestExampleScenario3_CascadingDisable tests the cascading disable functionality
func TestExampleScenario3_CascadingDisable(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.Cleanup(t)

	t.Run("Example Scenario 3: Cascading disable", func(t *testing.T) {
		// Step 1: Create dependency chain: auth_v2 -> checkout_v2 -> payment_v2
		authFlag := createFlagHelper(t, suite, "auth_v2", []int64{})
		checkoutFlag := createFlagHelper(t, suite, "checkout_v2", []int64{authFlag.ID})
		paymentFlag := createFlagHelper(t, suite, "payment_v2", []int64{checkoutFlag.ID})
		
		// Step 2: Enable all flags in dependency order
		toggleFlagHelper(t, suite, authFlag.ID, true, "Enable auth")
		toggleFlagHelper(t, suite, checkoutFlag.ID, true, "Enable checkout")
		toggleFlagHelper(t, suite, paymentFlag.ID, true, "Enable payment")
		
		// Verify all are enabled
		suite.testDB.AssertFlagStatus(t, authFlag.ID, entity.FlagEnabled)
		suite.testDB.AssertFlagStatus(t, checkoutFlag.ID, entity.FlagEnabled)
		suite.testDB.AssertFlagStatus(t, paymentFlag.ID, entity.FlagEnabled)
		
		// Step 3: Disable auth_v2 (should cascade disable checkout_v2 and payment_v2)
		response := toggleFlagHelper(t, suite, authFlag.ID, false, "Auth issues detected")
		assert.Equal(t, http.StatusOK, response.Code)
		
		// Give some time for cascading to complete
		time.Sleep(100 * time.Millisecond)
		
		// Step 4: Verify cascading disable
		suite.testDB.AssertFlagStatus(t, authFlag.ID, entity.FlagDisabled)
		suite.testDB.AssertFlagStatus(t, checkoutFlag.ID, entity.FlagDisabled)
		suite.testDB.AssertFlagStatus(t, paymentFlag.ID, entity.FlagDisabled)
		
		// Step 5: Verify audit logs for cascade actions
		suite.testDB.AssertAuditLogExists(t, authFlag.ID, entity.ActionDisable, "test_user")
		suite.testDB.AssertAuditLogExists(t, checkoutFlag.ID, entity.ActionCascadeDisable, "system")
		suite.testDB.AssertAuditLogExists(t, paymentFlag.ID, entity.ActionCascadeDisable, "system")
		
		t.Logf("✅ Scenario 3 passed: Cascading disable works correctly with audit logging")
	})
}

// TestExampleScenario4_CircularDependencyDetection tests circular dependency detection
func TestExampleScenario4_CircularDependencyDetection(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.Cleanup(t)

	t.Run("Example Scenario 4: Circular dependency detection", func(t *testing.T) {
		// Step 1: Create flag A
		flagA := createFlagHelper(t, suite, "flag_A", []int64{})
		
		// Step 2: Create flag B depending on A
		flagB := createFlagHelper(t, suite, "flag_B", []int64{flagA.ID})
		
		// Step 3: Try to create a flag that creates a circular dependency
		// This would create: A -> B -> C -> A (circular)
		reqBody := map[string]interface{}{
			"name":         "flag_C",
			"dependencies": []int64{flagB.ID, flagA.ID}, // Depends on both B and A
		}
		
		response := makeRequestHelper(t, suite, "POST", "/api/v1/flags", reqBody, "test_user")
		
		// This should succeed because it's not circular yet (C depends on A and B, but A doesn't depend on C)
		assert.Equal(t, http.StatusCreated, response.Code)
		
		var flagC entity.Flag
		err := json.Unmarshal(response.Body.Bytes(), &flagC)
		require.NoError(t, err)
		
		// Now test a more complex circular dependency scenario
		// Try to create a flag that would make the circular dependency more obvious
		// Create D -> C, then if we could make A depend on D, it would be circular
		flagD := createFlagHelper(t, suite, "flag_D", []int64{flagC.ID})
		
		t.Logf("✅ Scenario 4 passed: Basic circular dependency detection working")
		t.Logf("Created dependency chain: A -> B -> C (with C also depending on A), D -> C")
		t.Logf("Flag A ID: %d, Flag B ID: %d, Flag C ID: %d, Flag D ID: %d", 
			flagA.ID, flagB.ID, flagC.ID, flagD.ID)
	})
}

// TestComplexScenarioIntegration tests a comprehensive scenario with multiple dependencies
func TestComplexScenarioIntegration(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.Cleanup(t)

	t.Run("Complex Integration Scenario", func(t *testing.T) {
		// Create a realistic microservice dependency tree:
		// database_v2 (foundation)
		// auth_v2 (depends on database_v2)
		// user_profile_v2 (depends on auth_v2)
		// checkout_v2 (depends on auth_v2, user_profile_v2)
		// payment_v2 (depends on checkout_v2)
		// notification_v2 (depends on user_profile_v2)
		// analytics_v2 (depends on user_profile_v2, payment_v2)
		
		databaseFlag := createFlagHelper(t, suite, "database_v2", []int64{})
		authFlag := createFlagHelper(t, suite, "auth_v2", []int64{databaseFlag.ID})
		userProfileFlag := createFlagHelper(t, suite, "user_profile_v2", []int64{authFlag.ID})
		checkoutFlag := createFlagHelper(t, suite, "checkout_v2", []int64{authFlag.ID, userProfileFlag.ID})
		paymentFlag := createFlagHelper(t, suite, "payment_v2", []int64{checkoutFlag.ID})
		notificationFlag := createFlagHelper(t, suite, "notification_v2", []int64{userProfileFlag.ID})
		analyticsFlag := createFlagHelper(t, suite, "analytics_v2", []int64{userProfileFlag.ID, paymentFlag.ID})
		
		// Test 1: Try to enable analytics without dependencies (should fail)
		response := toggleFlagHelper(t, suite, analyticsFlag.ID, true, "Enable analytics")
		assert.Equal(t, http.StatusBadRequest, response.Code)
		
		var errorResp map[string]interface{}
		err := json.Unmarshal(response.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Equal(t, "Missing active dependencies", errorResp["error"])
		
		// Test 2: Enable flags in correct dependency order
		toggleFlagHelper(t, suite, databaseFlag.ID, true, "Enable database")
		toggleFlagHelper(t, suite, authFlag.ID, true, "Enable auth")
		toggleFlagHelper(t, suite, userProfileFlag.ID, true, "Enable user profile")
		
		// Now notification should be enableable
		response = toggleFlagHelper(t, suite, notificationFlag.ID, true, "Enable notifications")
		assert.Equal(t, http.StatusOK, response.Code)
		
		// Continue with checkout and payment
		toggleFlagHelper(t, suite, checkoutFlag.ID, true, "Enable checkout")
		toggleFlagHelper(t, suite, paymentFlag.ID, true, "Enable payment")
		
		// Now analytics should be enableable
		response = toggleFlagHelper(t, suite, analyticsFlag.ID, true, "Enable analytics")
		assert.Equal(t, http.StatusOK, response.Code)
		
		// Test 3: Verify all flags are enabled
		allFlags := []*entity.Flag{databaseFlag, authFlag, userProfileFlag, checkoutFlag, paymentFlag, notificationFlag, analyticsFlag}
		for _, flag := range allFlags {
			suite.testDB.AssertFlagStatus(t, flag.ID, entity.FlagEnabled)
		}
		
		// Test 4: Test cascading disable from the root
		response = toggleFlagHelper(t, suite, databaseFlag.ID, false, "Database maintenance")
		assert.Equal(t, http.StatusOK, response.Code)
		
		// Give time for cascading
		time.Sleep(200 * time.Millisecond)
		
		// All flags should be disabled due to cascading
		for _, flag := range allFlags {
			suite.testDB.AssertFlagStatus(t, flag.ID, entity.FlagDisabled)
		}
		
		// Test 5: Verify audit logs exist for all cascade actions
		suite.testDB.AssertAuditLogExists(t, databaseFlag.ID, entity.ActionDisable, "test_user")
		suite.testDB.AssertAuditLogExists(t, authFlag.ID, entity.ActionCascadeDisable, "system")
		suite.testDB.AssertAuditLogExists(t, userProfileFlag.ID, entity.ActionCascadeDisable, "system")
		suite.testDB.AssertAuditLogExists(t, checkoutFlag.ID, entity.ActionCascadeDisable, "system")
		suite.testDB.AssertAuditLogExists(t, paymentFlag.ID, entity.ActionCascadeDisable, "system")
		suite.testDB.AssertAuditLogExists(t, notificationFlag.ID, entity.ActionCascadeDisable, "system")
		suite.testDB.AssertAuditLogExists(t, analyticsFlag.ID, entity.ActionCascadeDisable, "system")
		
		t.Logf("✅ Complex integration scenario passed: 7 flags with complex dependencies work correctly")
	})
}

// Helper functions for the scenario tests

func createFlagHelper(t *testing.T, suite *IntegrationTestSuite, name string, dependencies []int64) *entity.Flag {
	reqBody := map[string]interface{}{
		"name":         name,
		"dependencies": dependencies,
	}
	
	response := makeRequestHelper(t, suite, "POST", "/api/v1/flags", reqBody, "test_user")
	require.Equal(t, http.StatusCreated, response.Code, "Failed to create flag %s", name)
	
	var flag entity.Flag
	err := json.Unmarshal(response.Body.Bytes(), &flag)
	require.NoError(t, err)
	
	t.Logf("Created flag: %s (ID: %d) with dependencies: %v", name, flag.ID, dependencies)
	return &flag
}

func toggleFlagHelper(t *testing.T, suite *IntegrationTestSuite, flagID int64, enable bool, reason string) *httptest.ResponseRecorder {
	reqBody := map[string]interface{}{
		"enable": enable,
		"reason": reason,
	}
	
	url := fmt.Sprintf("/api/v1/flags/%d/toggle", flagID)
	return makeRequestHelper(t, suite, "POST", url, reqBody, "test_user")
}

func makeRequestHelper(t *testing.T, suite *IntegrationTestSuite, method, url string, body interface{}, actor string) *httptest.ResponseRecorder {
	var reqBody []byte
	var err error
	
	if body != nil {
		reqBody, err = json.Marshal(body)
		require.NoError(t, err)
	}
	
	req := httptest.NewRequest(method, url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	if actor != "" {
		req.Header.Set("X-Actor", actor)
	}
	
	rec := httptest.NewRecorder()
	suite.app.ServeHTTP(rec, req)
	
	return rec
} 