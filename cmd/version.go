package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// Current version of the CLI. This should be updated when bumping versions
const CurrentVersion = "1.0.4"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of ship",
	Long:  `All software has versions. This is ship's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("SHIP CLI v%s\n", CurrentVersion)
		checkUpdate()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func checkUpdate() {
	fmt.Println("\nChecking for updates...")
	
	resp, err := http.Get("https://api.github.com/repos/SHIP-platform/ship-cli/releases/latest")
	if err != nil {
		fmt.Println("Failed to check for updates.")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Failed to check for updates (bad status code).")
		return
	}

	var result struct {
		TagName string `json:"tag_name"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("Failed to parse release info.")
		return
	}

	latestVersion := result.TagName
	if len(latestVersion) > 0 && latestVersion[0] == 'v' {
		latestVersion = latestVersion[1:]
	}

	if latestVersion != CurrentVersion {
		fmt.Printf("🚀 A new version of SHIP CLI is available! (v%s -> v%s)\n", CurrentVersion, latestVersion)
		
		fmt.Print("Would you like to update now? [Y/n]: ")
		var response string
		fmt.Scanln(&response)
		
		if response == "" || response == "y" || response == "Y" {
			fmt.Println("Updating...")
			updateCmd := exec.Command("bash", "-c", "curl -sL https://console.ship-platform.com/install.sh | bash")
			updateCmd.Stdout = os.Stdout
			updateCmd.Stderr = os.Stderr
			err := updateCmd.Run()
			if err != nil {
				fmt.Printf("Update failed: %v\n", err)
				fmt.Println("Please try running the install command manually:")
				fmt.Println("curl -sL https://console.ship-platform.com/install.sh | bash")
			} else {
				fmt.Println("Update successful!")
			}
		}
	} else {
		fmt.Println("You are running the latest version.")
	}
}
