package main

import (
	"bufio"
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"sort"
	"time"

	"github.com/mattn/go-tflite"
	"github.com/nfnt/resize"
)

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func top(a []float32) int {
	t := 0
	m := float32(0)
	for i, e := range a {
		if i == 0 || e > m {
			m = e
			t = i
		}
	}
	return t
}

func loadLabels(filename string) ([]string, error) {
	labels := []string{}
	f, err := os.Open("labels.txt")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		labels = append(labels, scanner.Text())
	}
	return labels, nil
}

func main() {
	defer timeTrack(time.Now(), "main")
	// os.Setenv("TF_CPP_MIN_VLOG_LEVEL", "2")
	// os.Setenv("TF_CPP_MIN_LOG_LEVEL", "2")

	var modelPath, labelPath, imagePath string
	flag.StringVar(&modelPath, "model", "mobilenet_quant_v1_224.tflite", "path to model file")
	flag.StringVar(&labelPath, "label", "labels.txt", "path to label file")
	flag.StringVar(&imagePath, "image", "peacock.png", "path to image file")
	flag.Parse()

	f, err := os.Open(imagePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Fprintln(os.Stderr, "Image decoded")

	labels, err := loadLabels(labelPath)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Fprintln(os.Stderr, "Labels loaded")

	model := tflite.NewModelFromFile(modelPath)
	if model == nil {
		log.Fatal("cannot load model")
	}
	defer model.Delete()
	// fmt.Fprintln(os.Stderr, "Model loaded")

	options := tflite.NewInterpreterOptions()
	options.SetNumThread(4)
	options.SetErrorReporter(func(msg string, user_data interface{}) {
		fmt.Println(msg)
	}, nil)
	defer options.Delete()

	interpreter := tflite.NewInterpreter(model, options)
	if interpreter == nil {
		log.Fatal("cannot create interpreter")
	}
	defer interpreter.Delete()
	// fmt.Fprintln(os.Stderr, "New interpreter created")

	status := interpreter.AllocateTensors()
	if status != tflite.OK {
		log.Fatal("allocate failed")
	}
	// fmt.Fprintln(os.Stderr, "Tensors allocated")

	input := interpreter.GetInputTensor(0)
	wantedHeight := input.Dim(1)
	wantedWidth := input.Dim(2)
	wantedChannels := input.Dim(3)
	wantedType := input.Type()

	// start := time.now()
	// defer timeTrack(time.Now(), "main")

	qp := input.QuantizationParams()
	log.Printf("width: %v, height: %v, type: %v, scale: %v, zeropoint: %v", wantedWidth, wantedHeight, input.Type(), qp.Scale, qp.ZeroPoint)
	log.Printf("input tensor count: %v, output tensor count: %v", interpreter.GetInputTensorCount(), interpreter.GetOutputTensorCount())
	if qp.Scale == 0 {
		qp.Scale = 1
	}

	// resized := resize.Resize(uint(wantedWidth), uint(wantedHeight), img, resize.NearestNeighbor)
	resized := resize.Resize(uint(wantedWidth), uint(wantedHeight), img, resize.Bicubic)
	bounds := resized.Bounds()
	dx, dy := bounds.Dx(), bounds.Dy()
	// fmt.Fprintln(os.Stderr, resized)

	if wantedType == tflite.Float32 {
		ff := make([]float32, dx*dy*wantedChannels)
		for y := 0; y < dy; y++ {
			for x := 0; x < dx; x++ {
				r, g, b, _ := resized.At(x, y).RGBA()
				// fmt.Fprintln(os.Stderr, r)
				// ff[(y*dx+x)*3+0] = float32(r) / 65535.0
				// ff[(y*dx+x)*3+1] = float32(g) / 65535.0
				// ff[(y*dx+x)*3+2] = float32(b) / 65535.0

				// Other (ImageNet ??)
				// ff[(y*dx+x)*3+0] = ((float32(r) / 255.0) - (0.485 * 255.0)) / (255.0 * 0.229)
				// ff[(y*dx+x)*3+1] = ((float32(g) / 255.0) - (0.456 * 255.0)) / (255.0 * 0.224)
				// ff[(y*dx+x)*3+2] = ((float32(b) / 255.0) - (0.456 * 255.0)) / (255.0 * 0.225)

				// fmt.Fprintln(os.Stderr, float32(b)/255)

				// Normalize to [0, 1] (Mobilenet)
				// ff[(y*dx+x)*3+0] = (float32(r) / 255.0) / 255.0
				// ff[(y*dx+x)*3+1] = (float32(g) / 255.0) / 255.0
				// ff[(y*dx+x)*3+2] = (float32(b) / 255.0) / 255.0

				// Normalize to [-1, 1]
				meanValue := float32(127.5)
				stdDevValue := float32(127.5)
				ff[(y*dx+x)*3+0] = ((float32(r) / 255.0) - meanValue) / stdDevValue
				ff[(y*dx+x)*3+1] = ((float32(g) / 255.0) - meanValue) / stdDevValue
				ff[(y*dx+x)*3+2] = ((float32(b) / 255.0) - meanValue) / stdDevValue

			}
		}
		input.CopyFromBuffer(ff)
	} else if wantedType == tflite.UInt8 {
		bb := make([]byte, dx*dy*wantedChannels)
		for y := 0; y < dy; y++ {
			for x := 0; x < dx; x++ {
				col := resized.At(x, y)
				r, g, b, _ := col.RGBA()
				bb[(y*dx+x)*3+0] = byte(float64(r) / 255.0)
				bb[(y*dx+x)*3+1] = byte(float64(g) / 255.0)
				bb[(y*dx+x)*3+2] = byte(float64(b) / 255.0)
				// bb[(y*wantedWidth+x)*3+0] = byte(((r / 255.0) - qp.ZeroPoint) * qp.Scale)
				// bb[(y*wantedWidth+x)*3+1] = byte(((g / 255.0) - qp.ZeroPoint) * qp.Scale)
				// bb[(y*wantedWidth+x)*3+2] = byte(((b / 255.0) - qp.ZeroPoint) * qp.Scale)
				// fmt.Fprintln(os.Stderr, byte(float64(r)/255.0))
				// fmt.Fprintln(os.Stderr, bb[0])
			}
		}
		input.CopyFromBuffer(bb)
	} else {
		log.Fatal("is not wanted type: only UInt8 or Float32 accepted")
	}

	// fmt.Fprintln(os.Stderr, "Invoking...")
	status = interpreter.Invoke()
	if status != tflite.OK {
		log.Fatal("invoke failed")
	}

	output := interpreter.GetOutputTensor(0)

	// fmt.Fprintln(os.Stderr, top(output.Float32s()))

	outputSize := output.Dim(output.NumDims() - 1)

	// b := make([]byte, outputSize)
	outb := make([]byte, outputSize)
	outf := make([]float32, outputSize)

	type result struct {
		score float64
		index int
	}

	// fmt.Fprintln(os.Stderr, "Output Tensors generated")

	// out := output.Float32s()
	// fmt.Fprintln(os.Stderr, out)
	minRange := float32(1)
	maxRange := float32(1)

	if wantedType == tflite.UInt8 {
		status = output.CopyToBuffer(&outb[0])
	} else {
		status = output.CopyToBuffer(&outf[0])

		sort.Slice(output.Float32s(), func(i, j int) bool {
			return output.Float32s()[i] > output.Float32s()[j]
		})
		minRange = output.Float32s()[outputSize-1]
		maxRange = output.Float32s()[0]
		fmt.Fprintln(os.Stderr, minRange)
		fmt.Fprintln(os.Stderr, maxRange)
	}
	if status != tflite.OK {
		log.Fatal("output failed")
	}

	results := []result{}
	for i := 0; i < outputSize; i++ {
		score := float64(0)
		if wantedType == tflite.UInt8 {
			score = float64(outb[i]) / 255.0
		} else {
			// score = float64(outf[i] / (maxRange - minRange))
			score = float64(outf[i])
		}
		// fmt.Fprintln(os.Stderr, score)

		if score < 0.2 {
			continue
		}
		results = append(results, result{score: score, index: i})
	}

	// end := time.now()

	// fmt.Fprintln(os.Stderr, "Sorting and fetching result...")
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})
	for i := 0; i < len(results); i++ {
		// fmt.Printf("%02d: %s: %f\n", results[i].index, labels[results[i].index], results[i].score)
		// fmt.Printf("{\"%s\":%f}", labels[results[i].index], results[i].score)
		fmt.Printf("{\"%s\":%f}\n", labels[results[i].index], results[i].score)
		if i == 0 {
			break
		}
	}
}
