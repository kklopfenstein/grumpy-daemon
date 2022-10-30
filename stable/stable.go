package stable

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

type PredictionsResp struct {
	Uuid string
}

type Prediction struct {
	Status string
	Output []string
	Error  string `json:"error"`
}

type PredictionStatusResult struct {
	Prediction Prediction
}

type Request struct {
	Inputs *Input `json:"inputs"`
}

type Input struct {
	Width             int64   `json:"width"`
	Height            int64   `json:"height"`
	NumOutputs        string  `json:"num_outputs"`
	GuidanceScale     float64 `json:"guidance_scale"`
	PromptStrength    float64 `json:"prompt_strength"`
	NumInferenceSteps int64   `json:"num_inference_steps"`
	Prompt            string  `json:"prompt"`
}

func Stable(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: StableGet(s, i),
		},
	})
}

func StableGet(s *discordgo.Session, i *discordgo.InteractionCreate) string {
	if i.Member == nil || i.Member.User == nil {
		return "Who are you?"
	}

	if len(i.ChannelID) == 0 {
		return "Where is this coming from?"
	}

	username := fmt.Sprintf("<@%s>", i.Member.User.ID)
	channelId := i.ChannelID

	options := i.ApplicationCommandData().Options

	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	input, err := createInputFromArgs(optionMap)

	if err != nil {
		return err.Error()
	}

	go runStable(s, username, channelId, input, i.Interaction)
	return fmt.Sprintf("Buildin' an image for \"%s\"", input.Prompt)
}

func createInputFromArgs(optionMap map[string]*discordgo.ApplicationCommandInteractionDataOption) (*Input, error) {
	input := Input{}

	if option, ok := optionMap["prompt"]; ok {
		input.Prompt = option.StringValue()
	} else {
		return nil, fmt.Errorf("no prompt")
	}

	if option, ok := optionMap["width"]; ok {
		input.Width = option.IntValue()
	} else {
		input.Width = 512
	}

	if option, ok := optionMap["height"]; ok {
		input.Height = option.IntValue()
	} else {
		input.Height = 512
	}

	if option, ok := optionMap["num_outputs"]; ok {
		input.NumOutputs = fmt.Sprint(option.IntValue())
	} else {
		input.NumOutputs = "1"
	}

	if option, ok := optionMap["guidance_scale"]; ok {
		input.GuidanceScale = option.FloatValue()
	} else {
		input.GuidanceScale = 7.5
	}

	if option, ok := optionMap["prompt_strength"]; ok {
		input.PromptStrength = option.FloatValue()
	} else {
		input.PromptStrength = 0.8
	}

	if option, ok := optionMap["num_inference_steps"]; ok {
		input.NumInferenceSteps = option.IntValue()
	} else {
		input.NumInferenceSteps = 50
	}

	return &input, nil
}

func runStable(s *discordgo.Session, username string, channelId string, input *Input, interaction *discordgo.Interaction) {
	image, err := callStableApi(input)

	if err != nil {
		s.FollowupMessageCreate(interaction, false, &discordgo.WebhookParams{
			Content: err.Error(),
		})
	} else {
		s.FollowupMessageCreate(interaction, false, &discordgo.WebhookParams{
			Content: image,
		})
	}
}

func callStableApi(input *Input) (string, error) {
	stableUrl := os.Getenv("STABLE_URL")
	if len(stableUrl) == 0 {
		return "", fmt.Errorf("stable_url not set")
	}
	// first get the CSRF
	resp, err := http.Get(stableUrl)
	if err != nil {
		return "", fmt.Errorf("failed to connect")
	}
	cookies := resp.Cookies()
	var csrf string
	for _, cookie := range cookies {
		if cookie.Name == "csrftoken" {
			csrf = cookie.Value
		}
	}

	if len(csrf) == 0 {
		return "", fmt.Errorf("failed to retrieve csrf")
	}

	// now use the CSRF to submit the job
	var payloadBuf bytes.Buffer
	request := Request{
		Inputs: input,
	}
	requestData, err := json.Marshal(&request)
	if err != nil {
		return "", fmt.Errorf("failed to construct request")
	}
	payloadBuf.Write(requestData)

	stableSubmitUrl := os.Getenv("STABLE_SUBMIT_URL")
	if len(stableSubmitUrl) == 0 {
		return "", fmt.Errorf("stable_submit_url not set")
	}
	resp, err = http.Post(stableSubmitUrl, "application/json", &payloadBuf)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Response code: %s", resp.Status)
		log.Printf("Body: '%s'", body)
		log.Fatal(err)
	}
	var predictionsResp PredictionsResp
	err = json.Unmarshal(body, &predictionsResp)
	if err != nil {
		log.Fatal(err)
	}
	uuid := predictionsResp.Uuid

	// use uuid to query for result
	tries := 0
	success := false
	var url string

	triesMax := 50
	for !success && tries < triesMax {
		stableStatusUrl := os.Getenv("STABLE_STATUS_URL")
		if len(stableStatusUrl) == 0 {
			return "", fmt.Errorf("stable_status_url not set")
		}
		resp, err = http.Get(fmt.Sprintf("%s/%s", stableStatusUrl, uuid))
		if err != nil {
			return "", fmt.Errorf("error connecting to status url")
		}

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		var predictionStatus PredictionStatusResult
		err = json.Unmarshal(body, &predictionStatus)
		if err != nil {
			return "", fmt.Errorf("error unmarshalling status response")
		}

		if predictionStatus.Prediction.Status == "succeeded" {
			success = true
			url = strings.Join(predictionStatus.Prediction.Output, "\n")
		} else if predictionStatus.Prediction.Status == "failed" {
			return "", fmt.Errorf("error: %s", predictionStatus.Prediction.Error)
		} else {
			time.Sleep(3 * time.Second)
		}
		tries++
	}

	if tries == triesMax && !success {
		return "", fmt.Errorf("maximum number of attempts")
	}

	if len(url) == 0 {
		return "", fmt.Errorf("no results")
	}

	return url, nil
}
