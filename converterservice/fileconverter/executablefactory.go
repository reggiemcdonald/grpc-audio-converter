package fileconverter

import encodings "github.com/reggiemcdonald/grpc-audio-converter/converterservice/enums"

const (
	ffmpeg      = "ffmpeg"
	formatFlag  = "-f"
	inputFlag   = "-i"
	mapFlag     = "-map"
	audioStream = "0:0"
)

// A command factory
type ExecutableFactory interface {
	// Creates the appropriate file conversion command
	// using the conversion attributes
	Build(job *ConversionAttributes) Executable
}

// The default executable factory implementation
type defaultExecutableFactory struct {}

func newDefaultExecutableFactory() ExecutableFactory {
	return &defaultExecutableFactory{}
}

/*
 * Returns a pointer to the command object
 */
func commandForDestEncoding(job *ConversionAttributes) Executable {
	return newDefaultExecutable(ffmpeg,
		formatFlag,
		job.Request.SourceEncoding.Name(),
		inputFlag,
		job.Request.SourceUrl,
		mapFlag,
		audioStream,
		formatFlag,
		job.Request.DestEncoding.Name(),
		job.TmpFile)
}

/*
 * Creates a command object for conversions to MP4.
 * Note: MPEG-4 is the container type, and M4A specifies audio only
 * so we force the extension to be the audio type
 */
func commandForMP4(job *ConversionAttributes) Executable {
	const m4a string = "m4a"
	job.TmpFile = newTempFilePath(job.Request.Id, m4a, job.Request.IncludeExtension)
	return commandForDestEncoding(job)
}

/*
 * Creates a command object for codecs that do not require special circumstances
 */
func defaultCommand(job *ConversionAttributes) Executable {
	job.TmpFile = newTempFilePath(job.Request.Id, job.Request.DestEncoding.Name(), job.Request.IncludeExtension)
	return commandForDestEncoding(job)
}

/*
 * Selects the appropriate command to be created
 */
func (e *defaultExecutableFactory) Build(job *ConversionAttributes) (cmd Executable) {
	switch job.Request.DestEncoding {
	case encodings.MP4:
		cmd = commandForMP4(job)
	default:
		cmd = defaultCommand(job)
	}
	return cmd
}




