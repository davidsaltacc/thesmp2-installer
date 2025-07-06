package main

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

//go:embed bundled.zip
var embeddedZip []byte

func extractZip(zipData []byte, targetDir string) error {
	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return fmt.Errorf("failed to read zip: %w", err)
	}

	for _, file := range reader.File {
		path := filepath.Join(targetDir, file.Name)

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(path, 0755); err != nil {
				return fmt.Errorf("failed to create dir: %w", err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("failed to create parent dir: %w", err)
		}

		srcFile, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in zip: %w", err)
		}
		defer srcFile.Close()

		dstFile, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			dstFile.Close()
			return fmt.Errorf("failed to copy file: %w", err)
		}
		dstFile.Close()
	}

	return nil
}

func loadMcSk(log *walk.TextEdit) {

	log.SetText(log.Text() + "Extracting files...")

	target := "C:\\Windows\\Temp\\thesmp2-installer-extracted"
	if err := extractZip(embeddedZip, target); err != nil {
		fmt.Errorf("Extraction failed:", err)
		return
	}

	log.SetText(log.Text() + "\r\nSuccessfully extracted")
	log.SetText(log.Text() + "\r\nDownloading fabric...")

	configDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}

	resp, err := http.Get("https://meta.fabricmc.net/v2/versions/loader/1.21.1/0.16.14/profile/json")
	if err != nil {
		panic(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	os.MkdirAll(configDir+"\\.minecraft\\versions\\fabric-loader-0.16.14-1.21.1", os.ModePerm)
	os.WriteFile(configDir+"\\.minecraft\\versions\\fabric-loader-0.16.14-1.21.1\\fabric-loader-0.16.14-1.21.1.json", body, 0644)

	log.SetText(log.Text() + "\r\nDownloaded fabric")
	log.SetText(log.Text() + "\r\nCreating launcher profile...")

	data, err := os.ReadFile(configDir + "\\.minecraft\\launcher_profiles.json")
	if err != nil {
		panic(err)
	}

	var root map[string]interface{}
	if err := json.Unmarshal(data, &root); err != nil {
		panic(err)
	}

	profiles, ok := root["profiles"].(map[string]interface{})
	if !ok {
		fmt.Println(`"profiles" is missing or not a JSON object`)
		return
	}

	now := time.Now().Format(time.RFC3339)

	profiles["thesmp2"] = map[string]interface{}{
		"created":       now,
		"lastUsed":      now,
		"name":          "THE SMP 2",
		"type":          "custom",
		"lastVersionId": "fabric-loader-0.16.14-1.21.1",
		"gameDir":       configDir + "\\.minecraft\\The_SMP_Instance",
		"icon":          "data:image/png;base64,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",
		"javaArgs":      "-Xmx4G -XX:+UnlockExperimentalVMOptions -XX:+UseG1GC -XX:G1NewSizePercent=20 -XX:G1ReservePercent=20 -XX:MaxGCPauseMillis=50 -XX:G1HeapRegionSize=32M",
		"memoryMax":     4096,
	}

	root["selectedProfile"] = "thesmp2"

	updatedData, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile(configDir+"\\.minecraft\\launcher_profiles.json", updatedData, 0644); err != nil {
		panic(err)
	}

	log.SetText(log.Text() + "\r\nCreated launcher profile")
	log.SetText(log.Text() + "\r\nMoving mods, shaders and resourcepacks")

	mvcmd := exec.Command("cmd", "/C", "move", target+"\\The_SMP_Instance", configDir+"\\.minecraft")
	mvcmd.Run() // only exis status is 1, indicating its already there, in that case we fuck off

	log.SetText(log.Text() + "\r\nMoved everything over")
	log.SetText(log.Text() + "\r\nFinished installing")
	log.SetText(log.Text() + "\r\nYou may close this window now. Your launcher will have a new profile with this modpack installed")

}

func main() {

	var clientComboBox *walk.ComboBox
	var mainWindow *walk.MainWindow

	MainWindow{
		AssignTo: &mainWindow,
		Title:    "TheSMP2 Installer",
		Size:     Size{1, 1}, // it rescales it perfectly then
		Layout:   VBox{},
		Children: []Widget{
			Label{
				Text: "Hello! This user-friendly (about as friendly as its creator) \nwill help you install this modpack on your device without you having to do pretty much anything. \nFirst of all, can I ask what minecraft launcher you are using? \nThis will help prepare the modpack.",
			},
			ComboBox{
				AssignTo: &clientComboBox,
				Model: []string{
					"Normal Minecraft Launcher",
					"SKlauncher",
					"Modrinth, Curseforge, Prism, MultiMC, or the likes",
				},
				CurrentIndex: 0,
			},
			PushButton{
				Text: "Continue",
				OnClicked: func() {
					if clientComboBox.CurrentIndex() == 0 || clientComboBox.CurrentIndex() == 1 {
						mw := MainWindow{
							AssignTo: &mainWindow,
							Title:    "TheSMP2 Installer",
							Size:     Size{1, 1},
							Layout:   VBox{},
							Children: []Widget{
								Label{
									Text: "Confirm to install the modpack for " + (func() string {
										if clientComboBox.CurrentIndex() == 0 {
											return "Minecraft Launcher"
										}
										return "SKlauncher"
									}()) + ".",
								},
								Composite{
									Layout: HBox{
										MarginsZero: true,
									},
									Children: []Widget{
										PushButton{
											Text: "Okay",
											OnClicked: func() {
												var log *walk.TextEdit
												var installBtn *walk.PushButton
												mw := MainWindow{
													AssignTo: &mainWindow,
													Title:    "TheSMP2 Installer",
													Size:     Size{400, 300},
													Layout:   VBox{},
													Children: []Widget{
														PushButton{
															AssignTo: &installBtn,
															Text:     "Press to start installing",
															OnClicked: func() {
																installBtn.SetEnabled(false)
																go loadMcSk(log)
															},
														},
														TextEdit{
															AssignTo: &log,
															ReadOnly: true,
														},
													},
												}
												mainWindow.Close()
												mw.Run()
											},
										},
										PushButton{
											Text:      "Cancel",
											OnClicked: func() { panic("Exited") },
										},
									},
								},
							},
						}
						mainWindow.Close()
						mw.Run()
					} else {
						var okButton *walk.PushButton
						var cancelButton *walk.PushButton
						var label *walk.Label
						mw := MainWindow{
							AssignTo: &mainWindow,
							Title:    "TheSMP2 Installer",
							Size:     Size{1, 1},
							Layout:   VBox{},
							Children: []Widget{
								Label{
									AssignTo: &label,
									Text:     "If you are using an advanced launcher, you should well know how to do this yourself.\nThis tool is primarily meant for technological cavemen.\nI'll just drop the instance files (mods, configs, shaders etc) on your Desktop,\nyou create a profile in your launcher (Fabric Loader 0.16.14, Minecraft 1.21.1),\nand copy over all the instance files yourself.\nThat sound good?",
								},
								Composite{
									Layout: HBox{
										MarginsZero: true,
									},
									Children: []Widget{
										PushButton{
											AssignTo: &okButton,
											Text:     "Okay",
											OnClicked: func() {
												okButton.SetEnabled(false)
												homeDir, err := os.UserHomeDir()
												if err != nil {
													panic(err)
												}
												if err := extractZip(embeddedZip, homeDir+"\\Desktop"); err != nil {
													panic(err)
												}
												cancelButton.SetText("Close")
												label.SetText("Alright, I dropped the files off at your desktop. The rest is your job.")
												mainWindow.SetSize(walk.Size{1, 1})
											},
										},
										PushButton{
											AssignTo:  &cancelButton,
											Text:      "Cancel",
											OnClicked: func() { panic("Exited") },
										},
									},
								},
							},
						}
						mainWindow.Close()
						mw.Run()
					}
				},
			},
		},
	}.Run()

}
