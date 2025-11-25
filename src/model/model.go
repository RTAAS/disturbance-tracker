package model

import (
	// DTrack
	"dtrack/log"

	// Standard
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"

	// 3rd-Party
	"github.com/mjibson/go-dsp/fft"
	"github.com/mjibson/go-dsp/window"
	"github.com/owulveryck/onnx-go"
	"github.com/owulveryck/onnx-go/backend/x/gorgonnx"
	"gorgonia.org/tensor"
)

const (
	SAMPLE_RATE        = 48000
	AUDIO_LENGTH_BYTES = 192000
	N_MELS             = 128
	N_FFT              = 2048
	HOP_LENGTH         = 512
	SPECTROGRAM_FRAMES = 188
	INT16_MAX          = 32768.0
)

// OnnxModel holds raw bytes AND the class labels.
type OnnxModel struct {
	RawBytes []byte
	Labels   []string
}

func Train() {
	log.Info("Model Training is handled by python -m ai.train")
}

func Load(model_path string) OnnxModel {
	log.Debug("Loading model from %s", model_path)

	bytes, err := os.ReadFile(model_path)
	if err != nil {
		log.Die("could not read ONNX file: %s", err)
	}

	ext := filepath.Ext(model_path)
	jsonPath := strings.TrimSuffix(model_path, ext) + "_labels.json"

	labelsBytes, err := os.ReadFile(jsonPath)
	if err != nil {
		log.Die("could not read Labels file: %s", err)
	}

	var labels []string
	if err := json.Unmarshal(labelsBytes, &labels); err != nil {
		log.Die("could not parse Labels JSON: %s", err)
	}

	log.Info("Model loaded with classes: %v", labels)

	return OnnxModel{RawBytes: bytes, Labels: labels}
}

func Prepare(pcmData []byte) (*tensor.Dense, error) {
	// 1. Pad/Truncate
	if len(pcmData) < AUDIO_LENGTH_BYTES {
		padding := make([]byte, AUDIO_LENGTH_BYTES-len(pcmData))
		pcmData = append(pcmData, padding...)
	}
	if len(pcmData) > AUDIO_LENGTH_BYTES {
		pcmData = pcmData[:AUDIO_LENGTH_BYTES]
	}

	// 2. Normalize
	audioFloat32 := normalizeAudio(pcmData)
	audioFloat64 := make([]float64, len(audioFloat32))
	for i, v := range audioFloat32 {
		audioFloat64[i] = float64(v)
	}

	// 3. STFT Calculation
	n_frames_calculated := (len(audioFloat64)-N_FFT)/HOP_LENGTH + 1
	if n_frames_calculated > SPECTROGRAM_FRAMES {
		n_frames_calculated = SPECTROGRAM_FRAMES
	}

	// Calculate bins - USING FULL SPECTRUM (Matches skimage resize behavior)
	bins_to_keep := N_FFT/2 + 1 // 1025 bins

	spectrogram := make([][]float64, bins_to_keep)
	for i := range spectrogram {
		spectrogram[i] = make([]float64, SPECTROGRAM_FRAMES)
	}

	window_func := window.Hann(N_FFT)

	for i := 0; i < n_frames_calculated; i++ {
		start := i * HOP_LENGTH
		end := start + N_FFT
		if end > len(audioFloat64) {
			end = len(audioFloat64)
		}

		segment := audioFloat64[start:end]
		windowed := make([]float64, N_FFT)
		for j := 0; j < len(segment); j++ {
			windowed[j] = segment[j] * window_func[j]
		}

		complex_input := make([]complex128, N_FFT)
		for j := 0; j < N_FFT; j++ {
			complex_input[j] = complex(windowed[j], 0)
		}
		complex_result := fft.FFT(complex_input)

		// Magnitude Square
		for j := 0; j < bins_to_keep; j++ {
			val := complex_result[j]
			magSq := real(val)*real(val) + imag(val)*imag(val)
			spectrogram[j][i] = magSq
		}
	}

	// 4. Flatten & Resize
	power_spec_1D := make([]float64, bins_to_keep*SPECTROGRAM_FRAMES)
	for i := 0; i < bins_to_keep; i++ {
		copy(power_spec_1D[i*SPECTROGRAM_FRAMES:(i+1)*SPECTROGRAM_FRAMES], spectrogram[i])
	}

	// Resize using Average Pooling (Matches skimage better for full spectrum)
	rescaled_spec := resizeRowsAvg(power_spec_1D, bins_to_keep, SPECTROGRAM_FRAMES, N_MELS)

	// 5. Power to DB & Normalize
	mel_spec_db := powerToDb(rescaled_spec)
	img_normalized := minMaxNormalize(mel_spec_db)

	// 6. Stack Channels
	size := len(img_normalized)
	img3Channel := make([]float32, size*3)
	copy(img3Channel[:size], img_normalized)
	copy(img3Channel[size:size*2], img_normalized)
	copy(img3Channel[size*2:size*3], img_normalized)

	shape := []int{1, 3, N_MELS, SPECTROGRAM_FRAMES}

	inputTensor := tensor.New(
		tensor.Of(tensor.Float32),
		tensor.WithShape(shape...),
		tensor.WithBacking(img3Channel),
	)

	return inputTensor, nil
}

func Infer(inferModel OnnxModel, preparedAudio *tensor.Dense) map[string]float64 {
	backend := gorgonnx.NewGraph()
	onnxModel := onnx.NewModel(backend)

	if err := onnxModel.UnmarshalBinary(inferModel.RawBytes); err != nil {
		log.Die("could not unmarshal ONNX model: %s", err)
	}

	onnxModel.SetInput(0, tensor.Tensor(preparedAudio))
	if err := backend.Run(); err != nil {
		log.Die("Inference failed: %v", err)
	}

	outputTensors, _ := onnxModel.GetOutputTensors()
	outputDense, ok := outputTensors[0].(*tensor.Dense)
	if !ok {
		log.Die("Output tensor is not a *tensor.Dense type.")
	}

	floatSlice := outputDense.Data().([]float32)
	logits := make([]float64, len(floatSlice))
	for i, v := range floatSlice {
		logits[i] = float64(v)
	}

	probs := softmax(logits)

	results := make(map[string]float64)
	for i, label := range inferModel.Labels {
		if i < len(probs) {
			results[label] = probs[i]
		}
	}

	return results
}

// --- Helper Functions ---

func normalizeAudio(pcmData []byte) []float32 {
	numSamples := len(pcmData) / 2
	audioArray := make([]float32, numSamples)
	for i := 0; i < numSamples; i++ {
		val := int16(uint16(pcmData[i*2]) | uint16(pcmData[i*2+1])<<8)
		audioArray[i] = float32(val) / INT16_MAX
	}
	return audioArray
}

func powerToDb(spec []float64) []float64 {
	dbSpec := make([]float64, len(spec))
	var maxPower float64 = 0.0
	for _, v := range spec {
		if v > maxPower {
			maxPower = v
		}
	}
	if maxPower < 1e-10 {
		maxPower = 1.0
	}

	for i, v := range spec {
		dbSpec[i] = 10.0 * math.Log10(math.Max(v/maxPower, 1e-10))
	}
	return dbSpec
}

func minMaxNormalize(dbSpec []float64) []float32 {
	var minVal float64 = math.MaxFloat64
	var maxVal float64 = -math.MaxFloat64
	for _, v := range dbSpec {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}
	rangeVal := maxVal - minVal
	if rangeVal < 1e-6 {
		rangeVal = 1e-6
	}

	img := make([]float32, len(dbSpec))
	for i, v := range dbSpec {
		img[i] = float32((v - minVal) / rangeVal)
	}
	return img
}

// resizeRowsAvg: Average Pooling (using Full Spectrum)
func resizeRowsAvg(input []float64, originalRows, cols, targetRows int) []float64 {
	if originalRows == targetRows {
		return input
	}
	output := make([]float64, targetRows*cols)
	ratio := float64(originalRows) / float64(targetRows)

	for c := 0; c < cols; c++ {
		for r_target := 0; r_target < targetRows; r_target++ {
			startRow := int(math.Floor(float64(r_target) * ratio))
			endRow := int(math.Ceil(float64(r_target+1) * ratio))
			if endRow > originalRows {
				endRow = originalRows
			}

			sum := 0.0
			count := 0
			for r_orig := startRow; r_orig < endRow; r_orig++ {
				sum += input[r_orig*cols+c]
				count++
			}
			if count > 0 {
				output[r_target*cols+c] = sum / float64(count)
			}
		}
	}
	return output
}

func softmax(logits []float64) []float64 {
	max := -math.MaxFloat64
	for _, v := range logits {
		if v > max {
			max = v
		}
	}
	sum := 0.0
	exps := make([]float64, len(logits))
	for i, v := range logits {
		exps[i] = math.Exp(v - max)
		sum += exps[i]
	}
	for i := range exps {
		exps[i] /= sum
	}
	return exps
}
