package model_test

import (
	// DTrack
	"dtrack/model"

	// Standard
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper function to create a dummy raw audio buffer for dimensions testing
func createTestAudioBuffer(fillData bool) []byte {
	buffer := make([]byte, model.SampleSize)
	if fillData {
		for i := 0; i < model.SampleSize; i++ {
			buffer[i] = byte(i % 256)
		}
	}
	return buffer
}

// TestPrepare_Dimensions: Ensures DSP output has the correct shape [1, 3, 128, 188].
func TestPrepare_Dimensions(t *testing.T) {
	// Arrange: Create a valid input buffer (Silence)
	inputBuffer := createTestAudioBuffer(false)

	// Act: Call the Prepare function
	preparedTensor, err := model.Prepare(inputBuffer)
	if err != nil {
		t.Fatalf("Prepare failed during DSP processing: %v", err)
	}

	// Assert: Check the final expected shape
	expectedShape := []int{1, 1, model.Nmels, model.SpectrogramFrames}
	actualShape := preparedTensor.Shape()

	// Check dimensions count
	if len(actualShape) != len(expectedShape) {
		t.Fatalf("Dimension count mismatch. Expected %d, got %d", len(expectedShape), len(actualShape))
	}

	// Check specific dimensions (Channels, Height, Width)
	if actualShape[1] != expectedShape[1] || actualShape[2] != expectedShape[2] || actualShape[3] != expectedShape[3] {
		t.Errorf("Shape mismatch. Expected %v, got %v", expectedShape, actualShape)
	}

	t.Logf("Prepare() successful: Shape %v matches expected %v.", actualShape, expectedShape)
}

// Test known "empty" sample against trained model
func TestInfer_Empty(t *testing.T) {
	testMultiClass_Inference(t, "test_empty.dat", "empty")
}

// Test known "small_dog" sample against trained model
func TestInfer_SmallDog(t *testing.T) {
	testMultiClass_Inference(t, "test_smalldog.dat", "small_dog")
}

// Test known "big_dog" sample against trained model
func TestInfer_BigDog(t *testing.T) {
	testMultiClass_Inference(t, "test_bigdog.dat", "big_dog")
}

// Test known "combined" sample against trained model
func TestInfer_Combo(t *testing.T) {
	testMultiClass_Inference(t, "test_combo.dat", "combination")
}

// TestMultiClass_Inference: Loads real files and verifies Multi-Class output.
func testMultiClass_Inference(t *testing.T, audioPath string, expectedResult string) {
	// Define paths for Real Artifacts
	onnxPath := "test_model.onnx"

	// Determine expected JSON path
	ext := filepath.Ext(onnxPath)
	jsonPath := strings.TrimSuffix(onnxPath, ext) + ".labels"

	// Ensure files exist
	if _, err := os.Stat(onnxPath); os.IsNotExist(err) {
		t.Errorf("Skipping Test: Model file %s not found.", onnxPath)
	}
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		t.Errorf("Skipping Test: Labels file %s not found.", jsonPath)
	}
	if _, err := os.Stat(audioPath); os.IsNotExist(err) {
		t.Errorf("Skipping Test: Audio file %s not found.", audioPath)
	}

	// Load Model (This will also load the JSON labels)
	myModel := model.Load(onnxPath)

	if len(myModel.Labels) < 2 {
		t.Errorf("Labels not loaded correctly. Found: %v", myModel.Labels)
	}
	t.Logf("Model loaded with classes: %v", myModel.Labels)

	// Read Audio File
	rawBytes, err := os.ReadFile(audioPath)
	if err != nil {
		t.Fatalf("Could not read audio file: %v", err)
	}

	// Prepare Audio (DSP)
	preparedTensor, err := model.Prepare(rawBytes)
	if err != nil {
		t.Fatalf("Model Prepare failed (DSP): %v", err)
	}

	// Infer (Returns Map[string]float64)
	results := model.Infer(myModel, preparedTensor)

	// Check if map is empty
	if len(results) == 0 {
		t.Fatal("Inference returned empty result map.")
	}

	// Check if all expected labels are present
	best_class := ""
	best_probability := 0.0
	for _, label := range myModel.Labels {
		if _, ok := results[label]; !ok {
			t.Errorf("Missing probability for class: %s", label)
		}
		if results[label] > best_probability {
			best_class = label
			best_probability = results[label]
		}
	}

	sum := 0.0
	for _, prob := range results {
		sum += prob
	}

	if sum < 0.99 || sum > 1.01 {
		t.Errorf("Softmax failure: Probabilities sum to %f, expected ~1.0", sum)
	}

	// Validate the correct labels was found
	if best_class != expectedResult {
		t.Errorf("Expected %s, but found %s", expectedResult, best_class)
	}
}
