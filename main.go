package main

import (
	"context"
	"fmt"
	"github.com/ai-mastering/aimastering-go"
	"github.com/urfave/cli"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"time"
)

var (
	gitTag string
	gitHash string
	buildTime string
	goVersion string
	goos string
	goarch string
)

func uploadAudio(client *aimastering.APIClient, auth context.Context, audioPath string) (int32) {
	// upload input audio
	if audioPath == "-" {
		audio, _, err := client.AudioApi.CreateAudio(auth, map[string]interface{}{
			"file":  os.Stdin,
		})
		if err != nil {
			log.Fatal(err)
		}
		return audio.Id
	} else {
		audioFile, err := os.Open(audioPath)
		if err != nil {
			log.Fatal(err)
		}
		defer audioFile.Close()

		audio, _, err := client.AudioApi.CreateAudio(auth, map[string]interface{}{
			"file":  audioFile,
		})
		if err != nil {
			log.Fatal(err)
		}
		return audio.Id
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "aimastering"
	app.EnableBashCompletion = true
	app.Version = gitTag + " " + gitHash + " " + goos + "/" + goarch + " " + buildTime + " " + goVersion
	app.HelpName = "aimastering"
	app.Copyright = "(c) 2019 Bakuage Co., Ltd."
	app.Usage = "AI Mastering API CLI client"

	var accessToken string
	var input string
	var output string
	var userAgent string

	userAgentFlag := cli.StringFlag{
		Name:        "user-agent",
		Usage:       "User agent text used for API request",
		Hidden:      false,
		Value:       "aimastering-cli",
		Destination: &userAgent,
	}

	accessTokenFlag := cli.StringFlag{
		Name:        "access-token",
		Usage:       "AI Mastering API Access Token (retrieved from https://aimastering.com/app/developer)",
		EnvVar:      "AIMASTERING_ACCESS_TOKEN",
		Hidden:      false,
		Value:       "",
		Destination: &accessToken,
	}

	inputFlag := cli.StringFlag{
		Name:        "input, i",
		Usage:       "Input audio file path. If - is specified, stdin is used.",
		Hidden:      false,
		Value:       "",
		Destination: &input,
	}

	outputFlag := cli.StringFlag{
		Name:        "output, o",
		Usage:       "Output audio file path. If - is specified, stdout is used.",
		Hidden:      false,
		Value:       "",
		Destination: &output,
	}

	app.Commands = []cli.Command{
		{
			Name:    "master",
			Usage:   "master an audio",
			UsageText: "aimastering master --input input.wav --output output.wav [command options]",
			HideHelp:     false,
			Action:  func(c *cli.Context) error {
				if accessToken == "" {
					log.Fatal("--access-token required")
				}
				if input == "" {
					log.Fatal("--input required")
				}
				if output == "" {
					log.Fatal("--output required")
				}

				cfg := aimastering.NewConfiguration()
				cfg.UserAgent = userAgent
				log.Printf("User agent:%s\n", userAgent)
				client := aimastering.NewAPIClient(cfg)
				auth := context.WithValue(context.Background(), aimastering.ContextAPIKey, aimastering.APIKey{
					Key: accessToken,
				})

				// upload input audio
				inputAudioId := uploadAudio(client, auth, input)
				log.Printf("The input audio was uploaded id %d\n", inputAudioId)

				// start the mastering
				masteringOptions := map[string]interface{}{
					"mode": "custom",
					"target_loudness": float32(c.Float64("target-loudness")),
					"target_loudness_mode": c.String("target-loudness-mode"),
					"mastering": float32(c.Float64("mastering-level")) > 0,
					"mastering_matching_level": float32(c.Float64("mastering-level")),
					"mastering_algorithm": c.String("mastering-algorithm"),
					"ceiling_mode": c.String("ceiling-mode"),
					"ceiling": float32(c.Float64("ceiling")),
					"bass_preservation": c.Bool("bass-preservation"),
					"preset": c.String("preset"),
					"noise_reduction": c.Bool("noise-reduction"),
					"low_cut_freq": float32(c.Float64("high-cut-freq")),
					"high_cut_freq": float32(c.Float64("high-cut-freq")),
					"sample_rate": c.Int("sample-rate"),
					"bit_depth": c.Int("bit-depth"),
					"output_format": c.Int("output-format"),
					"oversample": float32(c.Int("oversample")),
				}
				if c.String("reference") != "" {
					referenceAudioId := uploadAudio(client, auth, input)
					log.Printf("The reference audio was uploaded id %d\n", referenceAudioId)
					masteringOptions["reference_audio_id"] = referenceAudioId
				}

				var masteringOptionKeys []string
				for k := range masteringOptions {
					masteringOptionKeys = append(masteringOptionKeys, k)
				}
				sort.Strings(masteringOptionKeys)
				log.Printf("Mastering options\n")
				for _, k := range masteringOptionKeys {
					log.Printf("%s:%s\n", k, fmt.Sprint(masteringOptions[k]))
				}

				mastering, _, err := client.MasteringApi.CreateMastering(auth, inputAudioId, masteringOptions)

				if err != nil {
					log.Fatal(err)
				}
				log.Printf("The mastering started id %d\n", mastering.Id)

				// wait for the mastering completion
				for mastering.Status == "processing" || mastering.Status == "waiting" {
					mastering, _, err = client.MasteringApi.GetMastering(auth, mastering.Id)
					if err != nil {
						log.Fatal(err)
					}
					log.Printf("waiting for the mastering completion %d%%\n", int(100 * mastering.Progression))
					time.Sleep(5 * time.Second)
				}

				if mastering.Status != "succeeded" {
					additionalReason := ""
					if mastering.FailureReason == "failed_to_prepare" {
						inputAudio, _, err := client.AudioApi.GetAudio(auth, mastering.InputAudioId)
						if err != nil {
							log.Fatal(err)
						}
						if inputAudio.Status != "prepared" {
							additionalReason += fmt.Sprintf("Input audio preparation failed with status %s because %s", inputAudio.Status, inputAudio.FailureReason)
						}

						if c.String("reference") != "" {
							referenceAudio, _, err := client.AudioApi.GetAudio(auth, mastering.ReferenceAudioId)
							if err != nil {
								log.Fatal(err)
							}
							if referenceAudio.Status != "prepared" {
								additionalReason += fmt.Sprintf("Reference audio preparation failed with status %s because %s", referenceAudio.Status, referenceAudio.FailureReason)
							}
						}
					}
					log.Fatalf("Mastering failed with status %s because %s, %s", mastering.Status, mastering.FailureReason, additionalReason)
				}

				// download output audio
				// notes
				// - client.AudioApi.DownloadAudio cannot be used because swagger-codegen doesn't support binary string response in golang
				// - instead use GetAudioDownloadToken (to get signed url) + HTTP Get
				audioDownloadToken, _, err := client.AudioApi.GetAudioDownloadToken(auth, mastering.OutputAudioId)

				// http get signed url
				resp, err := http.Get(audioDownloadToken.DownloadUrl)
				if err != nil {
					log.Fatal(err)
				}
				defer resp.Body.Close()

				if output == "-" {
					// write output
					_, err = io.Copy(os.Stdout, resp.Body)
					if err != nil  {
						log.Fatal(err)
					}
					log.Print("The Output audio was saved to stdout\n")
				} else {
					outputAudioFile, err := os.Create(output)
					if err != nil  {
						log.Fatal(err)
					}
					defer outputAudioFile.Close()

					// write output
					_, err = io.Copy(outputAudioFile, resp.Body)
					if err != nil  {
						log.Fatal(err)
					}
					log.Printf("The Output audio was saved to %s\n", output)
				}

				outputVideo := c.String("output-video")
				if outputVideo != "" {
					// wait for the video encode completion
					for mastering.VideoStatus == "waiting" {
						mastering, _, err = client.MasteringApi.GetMastering(auth, mastering.Id)
						if err != nil {
							log.Fatal(err)
						}
						log.Print("waiting for the video encode completion\n")
						time.Sleep(5 * time.Second)
					}

					if mastering.VideoStatus != "succeeded" {
						log.Fatalf("Video encode failed with status %s", mastering.VideoStatus)
					}

					// download output video
					videoDownloadToken, _, err := client.VideoApi.GetVideoDownloadToken(auth, mastering.OutputVideoId)

					// http get signed url
					resp, err := http.Get(videoDownloadToken.DownloadUrl)
					if err != nil {
						log.Fatal(err)
					}
					defer resp.Body.Close()

					outputVideoFile, err := os.Create(outputVideo)
					if err != nil  {
						log.Fatal(err)
					}
					defer outputVideoFile.Close()

					// write output
					_, err = io.Copy(outputVideoFile, resp.Body)
					if err != nil  {
						log.Fatal(err)
					}
					log.Printf("The Output video was saved to %s\n", outputVideo)
				}

				if c.Bool("remove") {
					client.MasteringApi.DeleteMastering(auth, mastering.Id)
				}

				return nil
			},
			Flags: []cli.Flag{
				accessTokenFlag,
				inputFlag,
				outputFlag,
				userAgentFlag,
				cli.StringFlag{
					Name:        "reference",
					Usage:       "Reference audio file path",
					Hidden:      true,
					Value:       "",
				},
				cli.StringFlag{
					Name:        "output-video",
					Usage:       "Output video file path. Save video only when specified.",
					Hidden:      false,
					Value:       "",
				},
				cli.Float64Flag{
					Name:        "target-loudness",
					Usage:       "Target loudness in dB",
					Hidden:      false,
					Value:       -9,
				},
				cli.StringFlag{
					Name:        "target-loudness-mode",
					Usage:       "Target loudness mode loudness/rms/peak/youtube_loudness",
					Hidden:      false,
					Value:       "loudness",
				},
				cli.Float64Flag{
					Name:        "mastering-level",
					Usage:       "Mastering level in [0, 1]. 0 means disabled",
					Hidden:      false,
					Value:       0.5,
				},
				cli.StringFlag{
					Name:        "mastering-algorithm",
					Usage:       "Mastering algorithm v1/v2",
					Hidden:      false,
					Value:       "v2",
				},
				cli.Float64Flag{
					Name:        "ceiling",
					Usage:       "Output ceiling in dB",
					Hidden:      false,
					Value:       -0.5,
				},
				cli.StringFlag{
					Name:        "ceiling-mode",
					Usage:       "Output ceiling mode peak/true_peak/lowpass_true_peak",
					Hidden:      false,
					Value:       "true_peak",
				},
				cli.BoolTFlag{
					Name:        "bass-preservation",
					Usage:       "Bass preservation",
					Hidden:      true,
				},
				cli.StringFlag{
					Name:        "preset",
					Usage:       "Mastering preset",
					Hidden:      false,
					Value:       "generic",
				},
				cli.BoolFlag{
					Name:        "noise-reduction",
					Usage:       "Noise reduction",
					Hidden:      true,
				},
				cli.Float64Flag{
					Name:        "low-cut-freq",
					Usage:       "Low cut frequency in Hz",
					Hidden:      false,
					Value: 20,
				},
				cli.Float64Flag{
					Name:        "high-cut-freq",
					Usage:       "High cut frequency in Hz",
					Hidden:      false,
					Value: 20000,
				},
				cli.IntFlag{
					Name:        "sample-rate",
					Usage:       "Sample rate of output. 0 means same as the input.",
					Hidden:      false,
					Value:       0,
				},
				cli.IntFlag{
					Name:        "bit-depth",
					Usage:       "Bit depth of output. This is used only when output format is wav. 16/24/32",
					Hidden:      false,
					Value:       24,
				},
				cli.StringFlag{
					Name:        "output-format",
					Usage:       "Output format of output. wav/mp3",
					Hidden:      false,
					Value:       "wav",
				},
				cli.IntFlag{
					Name:        "oversample",
					Usage:       "Oversample factor 1/2",
					Hidden:      false,
					Value:       2,
				},
				cli.BoolTFlag{
					Name:        "remove",
					Usage:       "Remove the created mastering before finish.",
					Hidden:      false,
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
