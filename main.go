package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
)

var (
	err error
)

//go:embed index.html
var index embed.FS

func main() {
	v4l2 := v4l2Controller{videoDevice: "/dev/video0"}
	http.Handle("/", http.FileServer(http.FS(index)))

	http.HandleFunc("/parameters/exposure", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if r.Method == "PUT" {
			exposure, err := strconv.Atoi(r.FormValue("exposure"))
			if err != nil {
				errHtml := fmt.Sprintf("<span class=\"error\">値の変更に失敗しました: %v</span>", err)
				http.Error(w, errHtml, http.StatusBadRequest)
				return
			}

			err = v4l2.setExposure(exposure)
			if err != nil {
				errHtml := fmt.Sprintf("<span class=\"error\">値の変更に失敗しました: %v</span>", err)
				http.Error(w, errHtml, http.StatusInternalServerError)
				return
			}
			fmt.Fprintf(w, "<span class=\"success\">値の変更に成功しました: %d</span>", exposure)
			return
		}

		http.Error(w, "<span class=\"error\">Method not allowed</span>", http.StatusMethodNotAllowed)
	})

	err := http.ListenAndServe(":8888", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}

type v4l2Controller struct {
	videoDevice string
}

func (c v4l2Controller) setExposure(value int) error {
	return c.setCtrl("exposure_time_absolute", value)
}

func (c v4l2Controller) getExposure() (int, error) {
	return c.getCtrl("exposure_time_absolute")
}

func (c v4l2Controller) setCtrl(ctrl string, value int) error {
	opts := []string{
		"--set-ctrl", fmt.Sprintf("%s=%d", ctrl, value),
		"--device", c.videoDevice,
	}
	cmd := exec.Command("v4l2-ctl", opts...)
	return cmd.Run()
}

func (c v4l2Controller) getCtrl(ctrl string) (int, error) {
	opts := []string{
		"--get-ctrl", ctrl,
		"--device", c.videoDevice,
	}
	cmd := exec.Command("v4l2-ctl", opts...)
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	results := strings.Split(string(out), ": ")
	if len(results) != 2 {
		return 0, fmt.Errorf("unexpected output: %s", out)
	}

	return strconv.Atoi(results[1])
}
