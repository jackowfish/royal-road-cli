package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"royal-road-cli/internal/config"
	"royal-road-cli/internal/ui"
)

var rootCmd = &cobra.Command{
	Use:   "royal-road-cli",
	Short: "A CLI client for reading Royal Road novels",
	Long:  `A terminal-based interface for browsing and reading novels from royalroad.com`,
	Run: func(cmd *cobra.Command, args []string) {
		// Show interactive menu when no command is given
		menuModel := ui.NewMenuModel()
		p := tea.NewProgram(menuModel, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			log.Fatal(err)
		}
	},
}

var readCmd = &cobra.Command{
	Use:   "read [fiction-id]",
	Short: "Read a fiction by ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fictionID := args[0]
		
		p := tea.NewProgram(ui.NewReaderModel(fictionID), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			log.Fatal(err)
		}
	},
}

var browseCmd = &cobra.Command{
	Use:   "browse",
	Short: "Browse popular fictions",
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(ui.NewBrowseModel(), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			log.Fatal(err)
		}
	},
}

var continueCmd = &cobra.Command{
	Use:   "continue",
	Short: "Continue reading your last book",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}
		
		lastEntry := cfg.GetLastReadEntry()
		if lastEntry == nil {
			fmt.Println("No reading history found. Use 'royal-road-cli' to start reading.")
			os.Exit(1)
		}
		
		fmt.Printf("Continuing: %s by %s\n", lastEntry.FictionTitle, lastEntry.Author)
		fmt.Printf("Chapter %d/%d: %s\n\n", lastEntry.CurrentChapter+1, lastEntry.TotalChapters, lastEntry.ChapterTitle)
		
		readerModel := ui.NewReaderModel(lastEntry.FictionID)
		readerModel.SetStartChapter(lastEntry.CurrentChapter)
		
		p := tea.NewProgram(readerModel, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(readCmd)
	rootCmd.AddCommand(browseCmd)
	rootCmd.AddCommand(continueCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}