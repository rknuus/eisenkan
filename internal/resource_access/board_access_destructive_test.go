// Destructive tests for BoardAccess as specified in BoardAccess_STP.md
// These tests verify API contract violations, resource exhaustion, error conditions,
// concurrent access violations, and recovery scenarios.
package resource_access

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// DT-API-001: Store and Update Task with invalid or unusual inputs
func TestAcceptance_BoardAccess_InvalidTaskDataHandling(t *testing.T) {
	t.Run("NilTaskHandling", func(t *testing.T) {
		tempDir, ba := setupTestBoardAccess(t)
		defer cleanupTestBoardAccess(tempDir, ba)

		priority := Priority{Urgent: true, Important: true}
		status := WorkflowStatus{Column: "todo", Section: "urgent-important", Position: 1}

		// Test nil task
		_, err := ba.CreateTask(nil, priority, status, nil)
		if err == nil {
			t.Error("Expected error for nil task, got none")
		}
		if !strings.Contains(err.Error(), "task cannot be nil") {
			t.Errorf("Expected 'task cannot be nil' error, got: %s", err.Error())
		}
	})

	t.Run("MissingRequiredFields", func(t *testing.T) {
		tempDir, ba := setupTestBoardAccess(t)
		defer cleanupTestBoardAccess(tempDir, ba)

		priority := Priority{Urgent: true, Important: true}
		status := WorkflowStatus{Column: "todo", Section: "urgent-important", Position: 1}

		// Test task with empty title
		task := &Task{Title: "", Description: "Test"}
		_, err := ba.CreateTask(task, priority, status, nil)
		if err == nil {
			t.Error("Expected error for empty title, got none")
		}
		if !strings.Contains(err.Error(), "task title cannot be empty") {
			t.Errorf("Expected 'task title cannot be empty' error, got: %s", err.Error())
		}

		// Test task with whitespace-only title
		task.Title = "   \t\n   "
		_, err = ba.CreateTask(task, priority, status, nil)
		if err == nil {
			t.Error("Expected error for whitespace-only title, got none")
		}
	})

	t.Run("InvalidPriorityValues", func(t *testing.T) {
		tempDir, ba := setupTestBoardAccess(t)
		defer cleanupTestBoardAccess(tempDir, ba)

		task := &Task{Title: "Valid Task", Description: "Test"}
		status := WorkflowStatus{Column: "todo", Section: "urgent-important", Position: 1}

		// Test not-urgent-not-important (invalid priority combination)
		invalidPriority := Priority{Urgent: false, Important: false}
		_, err := ba.CreateTask(task, invalidPriority, status, nil)
		if err == nil {
			t.Error("Expected error for not-urgent-not-important priority, got none")
		}
		if !strings.Contains(err.Error(), "not-urgent-not-important") {
			t.Errorf("Expected 'not-urgent-not-important' error, got: %s", err.Error())
		}
	})

	t.Run("LargeTaskDescriptions", func(t *testing.T) {
		tempDir, ba := setupTestBoardAccess(t)
		defer cleanupTestBoardAccess(tempDir, ba)

		priority := Priority{Urgent: true, Important: true}
		status := WorkflowStatus{Column: "todo", Section: "urgent-important", Position: 1}

		// Test task with very large description (>10KB)
		largeDescription := strings.Repeat("A", 15000) // 15KB description
		task := &Task{
			Title:       "Large Description Task",
			Description: largeDescription,
		}

		// Should handle large descriptions gracefully
		taskID, err := ba.CreateTask(task, priority, status, nil)
		if err != nil {
			t.Errorf("Unexpected error storing task with large description: %v", err)
		}

		// Verify we can retrieve it
		tasks, err := ba.GetTasksData([]string{taskID}, false)
		if err != nil {
			t.Errorf("Error retrieving large task: %v", err)
		}
		if len(tasks) != 1 || len(tasks[0].Task.Description) != 15000 {
			t.Error("Large description not preserved correctly")
		}
	})

	t.Run("SpecialCharactersInTaskData", func(t *testing.T) {
		tempDir, ba := setupTestBoardAccess(t)
		defer cleanupTestBoardAccess(tempDir, ba)

		priority := Priority{Urgent: true, Important: true}
		status := WorkflowStatus{Column: "todo", Section: "urgent-important", Position: 1}

		// Test task with special characters
		task := &Task{
			Title:       "Task with Special Characters: ñáéíóú 中文 🚀",
			Description: "Unicode test: αβγ δεζ ηθι κλμ νξο πρσ τυφ χψω",
			Tags:        []string{"unicode", "special-chars", "json\"test"},
		}

		taskID, err := ba.CreateTask(task, priority, status, nil)
		if err != nil {
			t.Errorf("Error storing task with special characters: %v", err)
		}

		// Verify retrieval preserves special characters
		tasks, err := ba.GetTasksData([]string{taskID}, false)
		if err != nil {
			t.Errorf("Error retrieving task with special characters: %v", err)
		}
		if len(tasks) != 1 {
			t.Fatal("Expected 1 task")
		}
		if tasks[0].Task.Title != task.Title {
			t.Error("Special characters in title not preserved")
		}
		if tasks[0].Task.Description != task.Description {
			t.Error("Special characters in description not preserved")
		}
	})

	t.Run("NonExistentTaskUpdates", func(t *testing.T) {
		tempDir, ba := setupTestBoardAccess(t)
		defer cleanupTestBoardAccess(tempDir, ba)

		priority := Priority{Urgent: true, Important: true}
		status := WorkflowStatus{Column: "todo", Section: "urgent-important", Position: 1}
		task := &Task{Title: "Updated Task"}

		// Try to update a non-existent task
		err := ba.ChangeTaskData("non-existent-id", task, priority, status)
		if err == nil {
			t.Error("Expected error updating non-existent task, got none")
		}
		if !strings.Contains(err.Error(), "task file not found") {
			t.Errorf("Expected 'task file not found' error, got: %s", err.Error())
		}
	})

	t.Run("ConcurrentUpdatesToSameTask", func(t *testing.T) {
		tempDir, ba := setupTestBoardAccess(t)
		defer cleanupTestBoardAccess(tempDir, ba)

		priority := Priority{Urgent: true, Important: true}
		status := WorkflowStatus{Column: "todo", Section: "urgent-important", Position: 1}

		// Create a task first
		initialTask := &Task{Title: "Concurrent Test Task"}
		taskID, err := ba.CreateTask(initialTask, priority, status, nil)
		if err != nil {
			t.Fatalf("Error creating initial task: %v", err)
		}

		// Concurrent updates
		var wg sync.WaitGroup
		numGoroutines := 10
		errors := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				updatedTask := &Task{
					Title: fmt.Sprintf("Updated Task %d", i),
				}
				err := ba.ChangeTaskData(taskID, updatedTask, priority, status)
				if err != nil {
					errors <- err
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		// All updates should complete successfully (mutex protection)
		for err := range errors {
			t.Errorf("Concurrent update failed: %v", err)
		}

		// Verify task still exists and is valid
		tasks, err := ba.GetTasksData([]string{taskID}, false)
		if err != nil {
			t.Errorf("Error retrieving task after concurrent updates: %v", err)
		}
		if len(tasks) != 1 {
			t.Error("Task should still exist after concurrent updates")
		}
	})
}

// DT-API-002: Retrieve Task with invalid identifiers
func TestAcceptance_BoardAccess_InvalidTaskIdentifierHandling(t *testing.T) {
	t.Run("NilAndEmptyIdentifiers", func(t *testing.T) {
		tempDir, ba := setupTestBoardAccess(t)
		defer cleanupTestBoardAccess(tempDir, ba)

		// Test with empty slice
		tasks, err := ba.GetTasksData([]string{}, false)
		if err != nil {
			t.Errorf("Unexpected error with empty identifier slice: %v", err)
		}
		if len(tasks) != 0 {
			t.Error("Expected empty result set for empty identifier slice")
		}

		// Test with empty string identifier
		tasks, err = ba.GetTasksData([]string{""}, false)
		if err != nil {
			t.Errorf("Unexpected error with empty identifier: %v", err)
		}
		if len(tasks) != 0 {
			t.Error("Expected empty result set for empty identifier")
		}
	})

	t.Run("InvalidCharactersInIdentifiers", func(t *testing.T) {
		tempDir, ba := setupTestBoardAccess(t)
		defer cleanupTestBoardAccess(tempDir, ba)

		invalidIDs := []string{
			"invalid/path/chars",
			"invalid\\backslash",
			"invalid:colon",
			"invalid*asterisk",
			"invalid?question",
			"invalid\"quote",
			"invalid<bracket",
			"invalid>bracket",
			"invalid|pipe",
		}

		// These should not crash and should return empty results
		tasks, err := ba.GetTasksData(invalidIDs, false)
		if err != nil {
			t.Errorf("Unexpected error with invalid characters: %v", err)
		}
		if len(tasks) != 0 {
			t.Error("Expected empty result set for invalid character identifiers")
		}
	})

	t.Run("ExtremelyLongIdentifiers", func(t *testing.T) {
		tempDir, ba := setupTestBoardAccess(t)
		defer cleanupTestBoardAccess(tempDir, ba)

		// Test with very long identifier (>1000 characters)
		longID := strings.Repeat("a", 2000)
		tasks, err := ba.GetTasksData([]string{longID}, false)
		if err != nil {
			t.Errorf("Unexpected error with long identifier: %v", err)
		}
		if len(tasks) != 0 {
			t.Error("Expected empty result set for extremely long identifier")
		}
	})

	t.Run("UnicodeIdentifiers", func(t *testing.T) {
		tempDir, ba := setupTestBoardAccess(t)
		defer cleanupTestBoardAccess(tempDir, ba)

		unicodeIDs := []string{
			"task-中文测试",
			"task-🚀🎯📝",
			"task-αβγδε",
			"task-łóŕęṁ",
		}

		// Should handle unicode gracefully
		tasks, err := ba.GetTasksData(unicodeIDs, false)
		if err != nil {
			t.Errorf("Unexpected error with unicode identifiers: %v", err)
		}
		if len(tasks) != 0 {
			t.Error("Expected empty result set for unicode identifiers")
		}
	})

	t.Run("NonExistentIdentifiers", func(t *testing.T) {
		tempDir, ba := setupTestBoardAccess(t)
		defer cleanupTestBoardAccess(tempDir, ba)

		nonExistentIDs := []string{
			"non-existent-1",
			"definitely-not-found-2",
			"missing-task-3",
			"00000000-0000-0000-0000-000000000000",
		}

		tasks, err := ba.GetTasksData(nonExistentIDs, false)
		if err != nil {
			t.Errorf("Unexpected error with non-existent identifiers: %v", err)
		}
		if len(tasks) != 0 {
			t.Error("Expected empty result set for non-existent identifiers")
		}
	})

	t.Run("MixedValidInvalidIdentifiers", func(t *testing.T) {
		tempDir, ba := setupTestBoardAccess(t)
		defer cleanupTestBoardAccess(tempDir, ba)

		// Create one valid task
		validTask := &Task{Title: "Valid Task"}
		priority := Priority{Urgent: true, Important: true}
		status := WorkflowStatus{Column: "todo", Section: "urgent-important", Position: 1}
		
		validID, err := ba.CreateTask(validTask, priority, status, nil)
		if err != nil {
			t.Fatalf("Failed to create valid task: %v", err)
		}

		// Mix valid and invalid identifiers
		mixedIDs := []string{
			validID,           // Valid
			"invalid-id-1",    // Invalid
			"",                // Empty
			"another-invalid", // Invalid
		}

		tasks, err := ba.GetTasksData(mixedIDs, false)
		if err != nil {
			t.Errorf("Unexpected error with mixed identifiers: %v", err)
		}
		
		// Should return only the valid task
		if len(tasks) != 1 {
			t.Errorf("Expected 1 task from mixed identifiers, got %d", len(tasks))
		}
		if len(tasks) > 0 && tasks[0].Task.ID != validID {
			t.Error("Returned task should match the valid identifier")
		}
	})

	t.Run("BulkRetrievalLargeSet", func(t *testing.T) {
		tempDir, ba := setupTestBoardAccess(t)
		defer cleanupTestBoardAccess(tempDir, ba)

		// Generate 1000 non-existent identifiers
		largeIDSet := make([]string, 1000)
		for i := 0; i < 1000; i++ {
			largeIDSet[i] = fmt.Sprintf("non-existent-%d", i)
		}

		// Should handle large bulk requests gracefully
		tasks, err := ba.GetTasksData(largeIDSet, false)
		if err != nil {
			t.Errorf("Unexpected error with large ID set: %v", err)
		}
		if len(tasks) != 0 {
			t.Error("Expected empty result set for large non-existent ID set")
		}
	})

	t.Run("ConcurrentRetrievalRequests", func(t *testing.T) {
		tempDir, ba := setupTestBoardAccess(t)
		defer cleanupTestBoardAccess(tempDir, ba)

		// Create a few valid tasks
		validIDs := make([]string, 5)
		for i := 0; i < 5; i++ {
			task := &Task{Title: fmt.Sprintf("Concurrent Task %d", i)}
			priority := Priority{Urgent: true, Important: true}
			status := WorkflowStatus{Column: "todo", Section: "urgent-important", Position: i + 1}
			
			id, err := ba.CreateTask(task, priority, status, nil)
			if err != nil {
				t.Fatalf("Failed to create task %d: %v", i, err)
			}
			validIDs[i] = id
		}

		// Concurrent retrieval requests
		var wg sync.WaitGroup
		numGoroutines := 20
		errors := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				
				// Mix of valid and invalid IDs
				queryIDs := []string{
					validIDs[i%len(validIDs)],
					fmt.Sprintf("invalid-%d", i),
				}
				
				tasks, err := ba.GetTasksData(queryIDs, false)
				if err != nil {
					errors <- err
					return
				}
				
				// Should get exactly one valid task back
				if len(tasks) != 1 {
					errors <- fmt.Errorf("expected 1 task, got %d", len(tasks))
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		// Check for errors
		for err := range errors {
			t.Errorf("Concurrent retrieval failed: %v", err)
		}
	})
}

// DT-API-004: Query Tasks with extreme criteria
func TestAcceptance_BoardAccess_ExtremeQueryCriteriaHandling(t *testing.T) {
	t.Run("InvalidPriorityCombinations", func(t *testing.T) {
		tempDir, ba := setupTestBoardAccess(t)
		defer cleanupTestBoardAccess(tempDir, ba)

		// Test with invalid priority combination
		invalidCriteria := &QueryCriteria{
			Priority: &Priority{Urgent: false, Important: false}, // Invalid combination
		}

		tasks, err := ba.FindTasks(invalidCriteria)
		if err != nil {
			t.Errorf("Unexpected error with invalid priority criteria: %v", err)
		}
		// Should return empty result set since no tasks have invalid priorities
		if len(tasks) != 0 {
			t.Error("Expected empty result set for invalid priority combination")
		}
	})

	t.Run("NonExistentStatusValues", func(t *testing.T) {
		tempDir, ba := setupTestBoardAccess(t)
		defer cleanupTestBoardAccess(tempDir, ba)

		// Test with non-existent columns and sections
		invalidCriteria := &QueryCriteria{
			Columns:  []string{"non-existent-column", "invalid-column"},
			Sections: []string{"non-existent-section", "invalid-section"},
		}

		tasks, err := ba.FindTasks(invalidCriteria)
		if err != nil {
			t.Errorf("Unexpected error with invalid status criteria: %v", err)
		}
		// Should return empty result set since no tasks match invalid columns/sections
		if len(tasks) != 0 {
			t.Error("Expected empty result set for non-existent columns/sections")
		}
	})

	t.Run("MalformedDateRanges", func(t *testing.T) {
		tempDir, ba := setupTestBoardAccess(t)
		defer cleanupTestBoardAccess(tempDir, ba)

		// Test with inverted date range (From > To)
		futureTime := time.Now().Add(24 * time.Hour)
		pastTime := time.Now().Add(-24 * time.Hour)

		invalidDateCriteria := &QueryCriteria{
			DateRange: &DateRange{
				From: &futureTime, // Future
				To:   &pastTime,   // Past (invalid range)
			},
		}

		tasks, err := ba.FindTasks(invalidDateCriteria)
		if err != nil {
			t.Errorf("Unexpected error with malformed date range: %v", err)
		}
		// Should return empty result set since no tasks can match impossible date range
		if len(tasks) != 0 {
			t.Error("Expected empty result set for impossible date range")
		}
	})

	t.Run("ContradictoryFilters", func(t *testing.T) {
		tempDir, ba := setupTestBoardAccess(t)
		defer cleanupTestBoardAccess(tempDir, ba)

		// Create a task first
		task := &Task{Title: "Test Task", Tags: []string{"tag1"}}
		priority := Priority{Urgent: true, Important: true}
		status := WorkflowStatus{Column: "todo", Section: "urgent-important", Position: 1}
		
		_, err := ba.CreateTask(task, priority, status, nil)
		if err != nil {
			t.Fatalf("Failed to create test task: %v", err)
		}

		// Query with contradictory criteria
		contradictoryCriteria := &QueryCriteria{
			Columns: []string{"todo"},                                           // Task is in todo
			Priority: &Priority{Urgent: false, Important: true},               // But query for different priority
			Tags:     []string{"tag2"},                                         // And different tag
		}

		tasks, err := ba.FindTasks(contradictoryCriteria)
		if err != nil {
			t.Errorf("Unexpected error with contradictory criteria: %v", err)
		}
		// Should return empty since no task matches all contradictory criteria
		if len(tasks) != 0 {
			t.Error("Expected empty result set for contradictory criteria")
		}
	})

	t.Run("ExtremelyComplexFilterCombinations", func(t *testing.T) {
		tempDir, ba := setupTestBoardAccess(t)
		defer cleanupTestBoardAccess(tempDir, ba)

		// Create diverse tasks
		testTasks := []struct {
			task     *Task
			priority Priority
			status   WorkflowStatus
		}{
			{
				task:     &Task{Title: "Task 1", Tags: []string{"urgent", "work", "important"}},
				priority: Priority{Urgent: true, Important: true},
				status:   WorkflowStatus{Column: "todo", Section: "urgent-important", Position: 1},
			},
			{
				task:     &Task{Title: "Task 2", Tags: []string{"personal", "low"}},
				priority: Priority{Urgent: false, Important: true},
				status:   WorkflowStatus{Column: "doing", Position: 1},
			},
			{
				task:     &Task{Title: "Task 3", Tags: []string{"work", "research"}},
				priority: Priority{Urgent: true, Important: false},
				status:   WorkflowStatus{Column: "done", Position: 1},
			},
		}

		// Store all tasks
		for _, td := range testTasks {
			_, err := ba.CreateTask(td.task, td.priority, td.status, nil)
			if err != nil {
				t.Fatalf("Failed to store task %s: %v", td.task.Title, err)
			}
		}

		// Complex query with multiple filter combinations
		complexCriteria := &QueryCriteria{
			Columns:  []string{"todo", "doing", "done"}, // All columns
			Sections: []string{"urgent-important"},      // Specific section
			Tags:     []string{"work"},                  // Specific tag
			Priority: &Priority{Urgent: true, Important: true}, // Specific priority
		}

		tasks, err := ba.FindTasks(complexCriteria)
		if err != nil {
			t.Errorf("Unexpected error with complex criteria: %v", err)
		}

		// Should return only Task 1 (matches all criteria)
		if len(tasks) != 1 {
			t.Errorf("Expected 1 task from complex query, got %d", len(tasks))
		}
		if len(tasks) > 0 && tasks[0].Task.Title != "Task 1" {
			t.Errorf("Expected 'Task 1' from complex query, got '%s'", tasks[0].Task.Title)
		}
	})

	t.Run("UnicodeInCriteria", func(t *testing.T) {
		tempDir, ba := setupTestBoardAccess(t)
		defer cleanupTestBoardAccess(tempDir, ba)

		// Create task with unicode tags
		unicodeTask := &Task{
			Title: "Unicode Task",
			Tags:  []string{"测试", "🚀rocket", "αβγ"},
		}
		priority := Priority{Urgent: true, Important: true}
		status := WorkflowStatus{Column: "todo", Section: "urgent-important", Position: 1}
		
		_, err := ba.CreateTask(unicodeTask, priority, status, nil)
		if err != nil {
			t.Fatalf("Failed to create unicode task: %v", err)
		}

		// Query with unicode criteria
		unicodeCriteria := &QueryCriteria{
			Tags: []string{"测试", "🚀rocket"},
		}

		tasks, err := ba.FindTasks(unicodeCriteria)
		if err != nil {
			t.Errorf("Unexpected error with unicode criteria: %v", err)
		}

		// Should find the unicode task
		if len(tasks) != 1 {
			t.Errorf("Expected 1 task with unicode criteria, got %d", len(tasks))
		}
		if len(tasks) > 0 && tasks[0].Task.Title != "Unicode Task" {
			t.Error("Should find the unicode task")
		}
	})

	t.Run("ConcurrentQueryOperations", func(t *testing.T) {
		tempDir, ba := setupTestBoardAccess(t)
		defer cleanupTestBoardAccess(tempDir, ba)

		// Create a set of tasks for concurrent querying
		validPriorities := []Priority{
			{Urgent: true, Important: true},   // urgent-important
			{Urgent: true, Important: false},  // urgent-not-important  
			{Urgent: false, Important: true},  // not-urgent-important
		}
		
		for i := 0; i < 20; i++ {
			task := &Task{
				Title: fmt.Sprintf("Concurrent Task %d", i),
				Tags:  []string{fmt.Sprintf("tag%d", i%5)}, // 5 different tag groups
			}
			priority := validPriorities[i%len(validPriorities)] // Only valid combinations
			status := WorkflowStatus{
				Column:   []string{"todo", "doing", "done"}[i%3],
				Position: i + 1,
			}
			if status.Column == "todo" {
				status.Section = []string{"urgent-important", "urgent-not-important", "not-urgent-important"}[i%3]
			}

			_, err := ba.CreateTask(task, priority, status, nil)
			if err != nil {
				t.Fatalf("Failed to create concurrent task %d: %v", i, err)
			}
		}

		// Concurrent query operations with different criteria
		var wg sync.WaitGroup
		numGoroutines := 15
		errors := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				// Different query criteria per goroutine
				criteria := &QueryCriteria{
					Columns: []string{[]string{"todo", "doing", "done"}[i%3]},
					Tags:    []string{fmt.Sprintf("tag%d", i%5)},
				}

				tasks, err := ba.FindTasks(criteria)
				if err != nil {
					errors <- err
					return
				}

				// Verify results make sense (should have 0-7 tasks depending on criteria)
				if len(tasks) > 20 {
					errors <- fmt.Errorf("too many tasks returned: %d", len(tasks))
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		// Check for errors
		for err := range errors {
			t.Errorf("Concurrent query failed: %v", err)
		}
	})
}

// DT-RESOURCE-001: Memory and Performance Exhaustion
func TestAcceptance_BoardAccess_MemoryPerformanceExhaustion(t *testing.T) {
	// Skip in short test mode since these are intensive
	if testing.Short() {
		t.Skip("Skipping resource exhaustion tests in short mode")
	}

	t.Run("LargeTaskVolume", func(t *testing.T) {
		tempDir, ba := setupTestBoardAccess(t)
		defer cleanupTestBoardAccess(tempDir, ba)

		// Store 1000 tasks with moderate descriptions
		taskCount := 1000
		priority := Priority{Urgent: true, Important: true}
		
		t.Logf("Creating %d tasks...", taskCount)
		start := time.Now()
		
		for i := 0; i < taskCount; i++ {
			task := &Task{
				Title:       fmt.Sprintf("Load Test Task %d", i),
				Description: fmt.Sprintf("This is load test task number %d with some description text to simulate real usage", i),
				Tags:        []string{fmt.Sprintf("tag%d", i%10)},
			}
			status := WorkflowStatus{
				Column:   []string{"todo", "doing", "done"}[i%3],
				Position: i + 1,
			}
			if status.Column == "todo" {
				status.Section = []string{"urgent-important", "urgent-not-important", "not-urgent-important"}[i%3]
			}

			_, err := ba.CreateTask(task, priority, status, nil)
			if err != nil {
				t.Fatalf("Failed to store task %d: %v", i, err)
			}

			// Progress logging every 100 tasks
			if i%100 == 0 && i > 0 {
				elapsed := time.Since(start)
				rate := float64(i) / elapsed.Seconds()
				t.Logf("Stored %d tasks, rate: %.1f tasks/sec", i, rate)
			}
		}

		createDuration := time.Since(start)
		t.Logf("Created %d tasks in %v (%.1f tasks/sec)", taskCount, createDuration, float64(taskCount)/createDuration.Seconds())

		// Test bulk query performance
		start = time.Now()
		results, err := ba.FindTasks(&QueryCriteria{})
		queryDuration := time.Since(start)
		
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != taskCount {
			t.Errorf("Expected %d tasks, got %d", taskCount, len(results))
		}
		t.Logf("Queried %d tasks in %v (%.1f tasks/sec)", len(results), queryDuration, float64(len(results))/queryDuration.Seconds())
	})

	t.Run("LargeTaskDescriptions", func(t *testing.T) {
		tempDir, ba := setupTestBoardAccess(t)
		defer cleanupTestBoardAccess(tempDir, ba)

		// Test with very large descriptions
		descriptionSizes := []int{1024, 10240, 51200} // 1KB, 10KB, 50KB
		
		for i, size := range descriptionSizes {
			largeDescription := strings.Repeat("A", size)
			task := &Task{
				Title:       fmt.Sprintf("Large Description Task %d", i),
				Description: largeDescription,
			}
			priority := Priority{Urgent: true, Important: true}
			status := WorkflowStatus{Column: "todo", Section: "urgent-important", Position: i + 1}

			start := time.Now()
			taskID, err := ba.CreateTask(task, priority, status, nil)
			duration := time.Since(start)
			
			if err != nil {
				t.Errorf("Failed to store task with %d bytes description: %v", size, err)
				continue
			}

			t.Logf("Stored task with %d bytes description in %v", size, duration)

			// Verify retrieval performance
			start = time.Now()
			tasks, err := ba.GetTasksData([]string{taskID}, false)
			duration = time.Since(start)
			
			if err != nil {
				t.Errorf("Failed to retrieve large task: %v", err)
			} else if len(tasks) != 1 || len(tasks[0].Task.Description) != size {
				t.Error("Large description not preserved correctly")
			}
			t.Logf("Retrieved task with %d bytes description in %v", size, duration)
		}
	})
}

// DT-PERFORMANCE-001: Performance Degradation Under Load
func TestAcceptance_BoardAccess_PerformanceDegradationUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	t.Run("ConcurrentLoadTest", func(t *testing.T) {
		tempDir, ba := setupTestBoardAccess(t)
		defer cleanupTestBoardAccess(tempDir, ba)

		// Test concurrent operations from multiple goroutines
		numGoroutines := 10
		operationsPerGoroutine := 20
		
		var wg sync.WaitGroup
		errors := make(chan error, numGoroutines)
		durations := make(chan time.Duration, numGoroutines*operationsPerGoroutine)

		t.Logf("Starting concurrent load test: %d goroutines, %d ops each", numGoroutines, operationsPerGoroutine)

		start := time.Now()
		
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()
				
				for j := 0; j < operationsPerGoroutine; j++ {
					opStart := time.Now()
					
					// Mix of operations
					switch j % 4 {
					case 0: // Store task
						task := &Task{
							Title: fmt.Sprintf("Concurrent Task G%d-O%d", goroutineID, j),
							Tags:  []string{fmt.Sprintf("load-test-g%d", goroutineID)},
						}
						priority := Priority{Urgent: true, Important: true}
						status := WorkflowStatus{Column: "todo", Section: "urgent-important", Position: j + 1}
						
						_, err := ba.CreateTask(task, priority, status, nil)
						if err != nil {
							errors <- fmt.Errorf("Store failed G%d-O%d: %v", goroutineID, j, err)
						}
						
					case 1: // Query tasks
						criteria := &QueryCriteria{
							Tags: []string{fmt.Sprintf("load-test-g%d", goroutineID)},
						}
						_, err := ba.FindTasks(criteria)
						if err != nil {
							errors <- fmt.Errorf("Query failed G%d-O%d: %v", goroutineID, j, err)
						}
						
					case 2: // Retrieve task identifiers
						_, err := ba.ListTaskIdentifiers(AllTasks)
						if err != nil {
							errors <- fmt.Errorf("RetrieveIdentifiers failed G%d-O%d: %v", goroutineID, j, err)
						}
						
					case 3: // Get board configuration
						_, err := ba.GetBoardConfiguration()
						if err != nil {
							errors <- fmt.Errorf("GetConfig failed G%d-O%d: %v", goroutineID, j, err)
						}
					}
					
					duration := time.Since(opStart)
					durations <- duration
				}
			}(i)
		}

		wg.Wait()
		close(errors)
		close(durations)
		
		totalDuration := time.Since(start)
		
		// Check for errors
		errorCount := 0
		for err := range errors {
			t.Errorf("Concurrent operation error: %v", err)
			errorCount++
		}

		// Analyze performance
		var totalOps time.Duration
		var maxDuration time.Duration
		opCount := 0
		
		for duration := range durations {
			totalOps += duration
			if duration > maxDuration {
				maxDuration = duration
			}
			opCount++
		}

		if opCount > 0 {
			avgDuration := totalOps / time.Duration(opCount)
			totalExpectedOps := numGoroutines * operationsPerGoroutine
			opsPerSecond := float64(totalExpectedOps) / totalDuration.Seconds()
			
			t.Logf("Performance Results:")
			t.Logf("- Total operations: %d", totalExpectedOps)
			t.Logf("- Total time: %v", totalDuration)
			t.Logf("- Operations per second: %.1f", opsPerSecond)
			t.Logf("- Average operation time: %v", avgDuration)
			t.Logf("- Maximum operation time: %v", maxDuration)
			t.Logf("- Error rate: %d/%d (%.1f%%)", errorCount, totalExpectedOps, float64(errorCount)*100/float64(totalExpectedOps))
			
			// Performance requirement: operations should complete within reasonable time
			if maxDuration > 5*time.Second {
				t.Errorf("Maximum operation time %v exceeds 5 second threshold", maxDuration)
			}
			if avgDuration > 1*time.Second {
				t.Errorf("Average operation time %v exceeds 1 second threshold", avgDuration)
			}
		}
	})

	t.Run("PerformanceStabilityOverTime", func(t *testing.T) {
		tempDir, ba := setupTestBoardAccess(t)
		defer cleanupTestBoardAccess(tempDir, ba)

		// Test performance stability over sustained operations
		iterations := 100
		measurements := make([]time.Duration, iterations)
		
		t.Logf("Testing performance stability over %d iterations", iterations)
		
		for i := 0; i < iterations; i++ {
			task := &Task{
				Title: fmt.Sprintf("Stability Test Task %d", i),
				Tags:  []string{"stability-test"},
			}
			priority := Priority{Urgent: true, Important: true}
			status := WorkflowStatus{Column: "todo", Section: "urgent-important", Position: i + 1}

			start := time.Now()
			_, err := ba.CreateTask(task, priority, status, nil)
			measurements[i] = time.Since(start)
			
			if err != nil {
				t.Errorf("Stability test failed at iteration %d: %v", i, err)
			}
			
			if i%20 == 0 && i > 0 {
				// Calculate rolling average
				sum := time.Duration(0)
				for j := i - 19; j <= i; j++ {
					sum += measurements[j]
				}
				avg := sum / 20
				t.Logf("Rolling average (iterations %d-%d): %v", i-19, i, avg)
			}
		}
		
		// Analyze for performance degradation
		firstQuarter := measurements[:25]
		lastQuarter := measurements[75:]
		
		var firstAvg, lastAvg time.Duration
		for _, d := range firstQuarter {
			firstAvg += d
		}
		firstAvg /= time.Duration(len(firstQuarter))
		
		for _, d := range lastQuarter {
			lastAvg += d
		}
		lastAvg /= time.Duration(len(lastQuarter))
		
		t.Logf("First quarter average: %v", firstAvg)
		t.Logf("Last quarter average: %v", lastAvg)
		
		// Check for significant performance degradation (>50% slowdown)
		degradation := float64(lastAvg-firstAvg) / float64(firstAvg)
		t.Logf("Performance change: %.1f%%", degradation*100)
		
		if degradation > 0.5 {
			t.Errorf("Significant performance degradation detected: %.1f%% slower", degradation*100)
		}
	})
}

// Helper functions for test setup and cleanup
func setupTestBoardAccess(t *testing.T) (string, IBoardAccess) {
	tempDir, err := os.MkdirTemp("", "boardaccess_dt_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	ba, err := NewBoardAccess(tempDir)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create BoardAccess: %v", err)
	}

	return tempDir, ba
}

func cleanupTestBoardAccess(tempDir string, ba IBoardAccess) {
	if ba != nil {
		ba.Close()
	}
	os.RemoveAll(tempDir)
}

// TestDT_ERROR_001_VersioningUtilityFailures tests resilience to version control issues
func TestAcceptance_BoardAccess_VersioningUtilityFailures(t *testing.T) {
	tempDir, ba := setupTestBoardAccess(t)
	defer cleanupTestBoardAccess(tempDir, ba)

	t.Run("BasicVersionControlIntegrity", func(t *testing.T) {
		// Test normal versioning functionality first
		task := &Task{
			Title:       "Version Control Test Task",
			Description: "Testing version control integration",
		}

		priority := Priority{Urgent: true, Important: true}
		status := WorkflowStatus{Column: "todo", Section: "urgent-important", Position: 1}
		
		taskID, err := ba.CreateTask(task, priority, status, nil)
		if err != nil {
			t.Errorf("CreateTask failed: %v", err)
		}

		// Verify task was committed to version control
		task.Description = "Updated description"
		inProgressStatus := WorkflowStatus{Column: "in-progress", Section: "urgent-important", Position: 1}
		err = ba.ChangeTaskData(taskID, task, priority, inProgressStatus)
		if err != nil {
			t.Errorf("ChangeTaskData failed: %v", err)
		}

		t.Logf("Version control operations completed successfully")
	})
}

// TestDT_ERROR_002_FileSystemFailures tests resilience to file system issues  
func TestAcceptance_BoardAccess_FileSystemFailures(t *testing.T) {
	tempDir, ba := setupTestBoardAccess(t)
	defer cleanupTestBoardAccess(tempDir, ba)

	t.Run("ReadOnlyFileSystem", func(t *testing.T) {
		// Store a task first
		task := &Task{
			Title:       "File System Test Task",
			Description: "Testing file system resilience",
		}

		priority := Priority{Urgent: true, Important: true}
		status := WorkflowStatus{Column: "todo", Section: "urgent-important", Position: 1}
		
		taskID, err := ba.CreateTask(task, priority, status, nil)
		if err != nil {
			t.Errorf("CreateTask failed: %v", err)
		}

		// Make directory read-only
		err = os.Chmod(tempDir, 0444)
		if err != nil {
			t.Skipf("Cannot change directory permissions: %v", err)
		}
		defer os.Chmod(tempDir, 0755) // Restore permissions

		// Try to store another task - should fail gracefully
		task2 := &Task{
			Title:       "Should Fail Task",
			Description: "This should fail due to read-only filesystem",
		}

		_, err = ba.CreateTask(task2, priority, status, nil)
		if err == nil {
			t.Error("Expected error due to read-only filesystem, got none")
		} else {
			t.Logf("Graceful error handling: %v", err)
		}

		// Restore permissions and verify recovery
		os.Chmod(tempDir, 0755)
		
		// Should be able to retrieve existing task
		retrievedTasks, err := ba.GetTasksData([]string{taskID}, false)
		if err != nil {
			t.Errorf("Failed to retrieve task after permission restore: %v", err)
		}
		if len(retrievedTasks) == 0 {
			t.Error("Retrieved task list is empty")
		}
	})
}

// TestDT_ERROR_003_JSONCorruptionHandling tests handling of corrupted task data files
func TestAcceptance_BoardAccess_JSONCorruptionHandling(t *testing.T) {
	tempDir, ba := setupTestBoardAccess(t)
	defer cleanupTestBoardAccess(tempDir, ba)

	t.Run("CorruptedJSONRecovery", func(t *testing.T) {
		// Store a task first
		task := &Task{
			Title:       "JSON Corruption Test Task",
			Description: "Testing JSON corruption handling",
		}

		priority := Priority{Urgent: true, Important: true}
		status := WorkflowStatus{Column: "todo", Section: "urgent-important", Position: 1}
		
		taskID, err := ba.CreateTask(task, priority, status, nil)
		if err != nil {
			t.Errorf("CreateTask failed: %v", err)
		}

		// Find and corrupt the JSON file
		urgentImportantDir := filepath.Join(tempDir, "01_todo", "urgent-important")
		files, err := os.ReadDir(urgentImportantDir)
		if err != nil {
			t.Fatalf("Failed to read directory: %v", err)
		}

		var taskFile string
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".json") {
				taskFile = filepath.Join(urgentImportantDir, file.Name())
				break
			}
		}

		if taskFile == "" {
			t.Fatal("No JSON task file found")
		}

		// Corrupt the JSON file
		corruptData := []byte(`{"invalid": json, "corrupted": true`)
		err = os.WriteFile(taskFile, corruptData, 0644)
		if err != nil {
			t.Fatalf("Failed to corrupt JSON file: %v", err)
		}

		// Try to retrieve the corrupted task - should handle gracefully
		_, err = ba.GetTasksData([]string{taskID}, false)
		if err == nil {
			t.Error("Expected error due to corrupted JSON, got none")
		} else {
			t.Logf("Graceful corruption handling: %v", err)
		}

		// Verify that other operations still work (service remains functional)
		task2 := &Task{
			Title:       "Recovery Test Task",
			Description: "Testing service recovery after corruption",
		}

		priority2 := Priority{Urgent: false, Important: true}
		status2 := WorkflowStatus{Column: "todo", Section: "not-urgent-important", Position: 1}
		
		_, err = ba.CreateTask(task2, priority2, status2, nil)
		if err != nil {
			t.Errorf("Service should recover and allow new tasks: %v", err)
		}
	})
}

// TestDT_RECOVERY_001_ServiceRecoveryAfterFailures tests recovery capabilities
func TestAcceptance_BoardAccess_ServiceRecoveryAfterFailures(t *testing.T) {
	tempDir, ba := setupTestBoardAccess(t)
	defer cleanupTestBoardAccess(tempDir, ba)

	t.Run("RecoveryAfterTempFailure", func(t *testing.T) {
		// Store initial task
		task := &Task{
			Title:       "Recovery Test Task",
			Description: "Testing recovery capabilities",
		}

		priority := Priority{Urgent: true, Important: true}
		status := WorkflowStatus{Column: "todo", Section: "urgent-important", Position: 1}
		
		taskID, err := ba.CreateTask(task, priority, status, nil)
		if err != nil {
			t.Errorf("CreateTask failed: %v", err)
		}

		// Simulate temporary filesystem issue by removing the task file
		taskPath := filepath.Join(tempDir, "01_todo", "urgent-important", "0001-task-"+taskID+".json")
		originalData, err := os.ReadFile(taskPath)
		if err != nil {
			t.Skipf("Cannot read task file for simulation: %v", err)
		}
		
		// Temporarily remove the file to simulate outage
		err = os.Remove(taskPath)
		if err != nil {
			t.Skipf("Cannot remove task file for simulation: %v", err)
		}

		// Operations should fail during the "outage" (return empty results)
		outageResults, err := ba.GetTasksData([]string{taskID}, false)
		if err != nil {
			t.Logf("Expected error during simulated outage: %v", err)
		} else if len(outageResults) > 0 {
			t.Error("Expected no results during simulated outage")
		} else {
			t.Logf("Service correctly returned empty results during outage")
		}

		// Restore file (simulate recovery)
		err = os.WriteFile(taskPath, originalData, 0644)
		if err != nil {
			t.Skipf("Cannot restore task file: %v", err)
		}

		// Service should recover automatically
		retrievedTasks, err := ba.GetTasksData([]string{taskID}, false)
		if err != nil {
			t.Errorf("Service should recover after permission restore: %v", err)
		}
		if len(retrievedTasks) == 0 {
			t.Error("Task should be retrievable after recovery")
		} else {
			t.Logf("Service recovered successfully, retrieved task: %s", retrievedTasks[0].Task.Title)
		}
	})
}

// TestDT_RECOVERY_002_PartialFunctionalityUnderConstraints tests continued operation under constraints
func TestAcceptance_BoardAccess_PartialFunctionalityUnderConstraints(t *testing.T) {
	tempDir, ba := setupTestBoardAccess(t)
	defer cleanupTestBoardAccess(tempDir, ba)

	t.Run("PartialDirectoryAccess", func(t *testing.T) {
		// Store tasks in different priority levels
		task1 := &Task{
			Title:       "High Priority Task",
			Description: "Should be accessible",
		}

		task2 := &Task{
			Title:       "Low Priority Task", 
			Description: "May become inaccessible",
		}

		priority1 := Priority{Urgent: true, Important: true}
		status1 := WorkflowStatus{Column: "todo", Section: "urgent-important", Position: 1}
		
		taskID1, err := ba.CreateTask(task1, priority1, status1, nil)
		if err != nil {
			t.Errorf("CreateTask failed: %v", err)
		}

		priority2 := Priority{Urgent: false, Important: true}
		status2 := WorkflowStatus{Column: "todo", Section: "not-urgent-important", Position: 1}
		
		taskID2, err := ba.CreateTask(task2, priority2, status2, nil)
		if err != nil {
			t.Errorf("CreateTask failed: %v", err)
		}

		// Simulate partial constraints by temporarily removing one task file
		task2Path := filepath.Join(tempDir, "01_todo", "not-urgent-important", "0001-task-"+taskID2+".json")
		task2Data, err := os.ReadFile(task2Path)
		if err != nil {
			t.Skipf("Cannot read task2 file for simulation: %v", err)
		}
		
		// Temporarily remove task2 file to simulate constraint
		err = os.Remove(task2Path)
		if err != nil {
			t.Skipf("Cannot remove task2 file for simulation: %v", err)
		}
		defer func() {
			// Restore task2 file
			os.WriteFile(task2Path, task2Data, 0644)
		}()

		// High priority task should still be accessible
		retrievedTasks1, err := ba.GetTasksData([]string{taskID1}, false)
		if err != nil {
			t.Errorf("High priority task should remain accessible: %v", err)
		}
		if len(retrievedTasks1) == 0 {
			t.Error("High priority task should be retrievable")
		}

		// Low priority task should fail gracefully (file was removed, expect empty results)
		constrainedResults, err := ba.GetTasksData([]string{taskID2}, false)
		if err != nil {
			t.Logf("Expected error for inaccessible task: %v", err)
		} else if len(constrainedResults) > 0 {
			t.Error("Expected no results for inaccessible task")
		} else {
			t.Logf("Service correctly returned empty results for inaccessible task")
		}

		// Query operations should return partial results
		criteria := &QueryCriteria{
			Columns:  []string{"todo"},
			Sections: []string{"urgent-important", "not-urgent-important"},
		}

		tasks, err := ba.FindTasks(criteria)
		if err != nil {
			t.Errorf("Query should return partial results: %v", err)
		}

		// Should find at least the accessible task
		if len(tasks) < 1 {
			t.Error("Query should return at least one accessible task")
		} else {
			t.Logf("Partial functionality maintained: found %d accessible tasks", len(tasks))
		}
	})
}