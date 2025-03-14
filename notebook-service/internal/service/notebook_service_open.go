package service

import (
	"fmt"
	"os/exec"
	"runtime"
)

func openURL(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Run()
}

var CallOpen = func(notebookname string, autoOpen bool) {

	environmentConfig := GetConfiguration()

	//build the link to the notebook
	notebooknameinput := notebookname

	notebookURL := environmentConfig.UrlBase + "/notebook/" + environmentConfig.Namespace + "/" + notebooknameinput + "/"
	if autoOpen {
		err := openURL(notebookURL)
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Printf("the link to the notebook is: " + notebookURL + "\n")
	}

}
