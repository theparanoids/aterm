package appdialogs

import (
	"os"

	"github.com/theparanoids/ashirt-server/backend/dtos"
	"github.com/theparanoids/aterm/cmd/aterm/config"
	"github.com/theparanoids/aterm/cmd/aterm/recording"
	"github.com/theparanoids/aterm/dialog"
	"github.com/theparanoids/aterm/fancy"
	"github.com/theparanoids/aterm/network"
)

func renderMainMenu(state MenuState) MenuState {
	rtnState := state
	menuOptions := []dialog.SimpleOption{
		dialogOptionStartRecording,
		dialogOptionUpdateOps,
		dialogOptionTestConnection,
		dialogOptionEditRunningConfig,
		dialogOptionExit,
	}

	resp := HandlePlainSelect("What do you want to do", menuOptions, func() dialog.SimpleOption {
		printline("Exiting...")
		return dialogOptionExit
	})
	switch {
	case dialogOptionStartRecording == resp.Selection:
		rtnState.CurrentView = MenuViewRecording

	case dialogOptionExit == resp.Selection:
		rtnState.CurrentView = MenuViewExit

	case dialogOptionTestConnection == resp.Selection:
		testConnection()

	case dialogOptionUpdateOps == resp.Selection:
		newOps, err := updateOperations()
		if err != nil {
			printline(fancy.Caution("Unable to retrieve operations list", err))
		} else {
			rtnState.AvailableOperations = newOps
		}

	case dialogOptionEditRunningConfig == resp.Selection:
		newConfig := editConfig(state.InstanceConfig)
		rtnState.InstanceConfig = newConfig
	default:
		printline("Hmm, I don't know how to handle that request. This is probably a bug. Could you please report this?")
	}

	return rtnState
}

func startNewRecording(state MenuState) MenuState {
	rtnState := state

	// collect info
	if len(state.AvailableOperations) == 0 {
		printline(fancy.ClearLine("Unable to record: No operations available (Try refreshing operations)"))
		rtnState.CurrentView = MenuViewMainMenu
		return rtnState
	}

	resp := askForOperationSlug(state.AvailableOperations, state.InstanceConfig.OperationSlug)

	recordedMetadata := RecordingMetadata{
		OperationSlug: unwrapOpSlug(resp),
	}
	// rtnState.InstanceConfig.OperationSlug = opSlug

	// reuse last tags, if they match the operation
	if recordedMetadata.OperationSlug == state.RecordedMetadata.OperationSlug {
		recordedMetadata.SelectedTags = state.RecordedMetadata.SelectedTags
	} else {
		recordedMetadata.SelectedTags = []dtos.Tag{}
	}

	rtnState.RecordedMetadata = recordedMetadata

	// start the recording
	rtnState.DialogInput = recording.DialogReader()
	output, err := recording.StartRecording(rtnState.RecordedMetadata.OperationSlug)

	if err != nil {
		printline(fancy.Fatal("Unable to record", err))
		rtnState.CurrentView = MenuViewMainMenu
		return rtnState
	}
	rtnState.RecordedMetadata.FilePath = output.FilePath
	rtnState.CurrentView = MenuViewUploadMenu

	return rtnState
}

func testConnection() {
	var testErr error
	var value string
	dialog.DoBackgroundLoading(
		dialog.SyncedFunc(func() {
			value, testErr = network.TestConnection()
		}),
	)
	if testErr != nil {
		printfln("%v Could not connect: %v", fancy.RedCross(), fancy.WithBold(testErr.Error(), fancy.Red))
		if value != "" {
			printline("Recommendation: " + value)
		}
		return
	}
	printfln("%v Connected", fancy.GreenCheck())
}

func updateOperations() ([]dtos.Operation, error) {
	var loadingErr error
	var ops []dtos.Operation
	dialog.DoBackgroundLoadingWithMessage("Retriving operations",
		dialog.SyncedFunc(func() {
			ops, loadingErr = network.GetOperations()
		}),
	)

	if loadingErr != nil {
		return []dtos.Operation{}, loadingErr
	}

	printf("Updated operations (%v total)\n", len(ops))
	return ops, nil
}

func editConfig(runningConfig config.TermRecorderConfig) config.TermRecorderConfig {
	rtnConfig := runningConfig
	overrideCfg := config.CloneConfigAsOverrides(runningConfig)

	// iterate through each question. After each, check if the user backed out via ^d/^c, and if so, stop asking questions and leave the function
	type FillQuestion struct {
		Fields     AskForTemplateFields
		DefaultVal *string
		AssignTo   *string
	}
	questions := []FillQuestion{
		FillQuestion{AssignTo: overrideCfg.AccessKey, Fields: accessKeyFields, DefaultVal: overrideCfg.AccessKey},
		FillQuestion{AssignTo: overrideCfg.SecretKey, Fields: secretKeyFields, DefaultVal: overrideCfg.SecretKey},
		FillQuestion{AssignTo: overrideCfg.APIURL, Fields: apiURLFields, DefaultVal: overrideCfg.APIURL},

		FillQuestion{AssignTo: overrideCfg.RecordingShell, Fields: shellFields, DefaultVal: thisOrThat(overrideCfg.RecordingShell, os.Getenv("SHELL"))},
		FillQuestion{AssignTo: overrideCfg.OutputDir, Fields: savePathFields, DefaultVal: overrideCfg.OutputDir},
	}

	stop := false
	for _, question := range questions {
		if !stop {
			*question.AssignTo = *askFor(question.Fields, question.DefaultVal, func() { stop = true }).Value
		}
	}
	if !stop {
		resp := askForOperationSlug(internalMenuState.AvailableOperations, runningConfig.OperationSlug)
		if resp.IsKillSignal() {
			stop = true
		} else {
			slug := unwrapOpSlug(resp)
			overrideCfg.OperationSlug = &slug
		}
	}

	if stop {
		printline("Discarding changes...")
		return rtnConfig
	}

	newCfg := config.PreviewUpdatedInstanceConfig(runningConfig, overrideCfg)

	config.PrintConfigTo(newCfg, medium)
	err := config.ValidateConfig(newCfg)
	if err != nil {
		ShowInvalidConfigMessageNoHelp(err)
	}
	yesPermanently := dialog.SimpleOption{Label: "Yes, and save for next time"}
	yesTemporarily := dialog.SimpleOption{Label: "Yes, for now"}
	cancelSave := dialog.SimpleOption{Label: "Cancel"}

	saveChangesOptions := []dialog.SimpleOption{
		yesPermanently,
		yesTemporarily,
		cancelSave,
	}
	resp := HandlePlainSelect("Do you want to save these changes", saveChangesOptions, func() dialog.SimpleOption {
		printline("Discarding changes...")
		return cancelSave
	})

	switch {
	case yesPermanently == resp.Selection:
		config.SetConfig(newCfg)
		if err := config.WriteConfig(); err != nil {
			ShowUnableToSaveConfigErrorMessage(err)
		}
		fallthrough
	case yesTemporarily == resp.Selection:
		network.SetAccessKey(newCfg.AccessKey)
		network.SetSecretKey(newCfg.SecretKey)
		network.SetBaseURL(newCfg.APIURL)
		rtnConfig = newCfg

	case cancelSave == resp.Selection:
		break

	default:
		printline("Hmm, I don't know how to handle that request. This is probably a bug. Could you please report this?")
	}

	return rtnConfig
}

func unwrapOpSlug(selectOpResp dialog.SelectResponse) string {
	if op, ok := selectOpResp.Selection.Data.(dtos.Operation); ok {
		return op.Slug
	}
	return ""
}
