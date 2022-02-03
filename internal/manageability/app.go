package manageability

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var strp string
var boolp bool
var intp int

var flagsCmd = &cobra.Command{
	Use:  "flags",
	Long: "A simple flags experimentation 1",
	Run:  flagsFunc,
}

var rootCmd = &cobra.Command{
	Use:  "cng",
	Long: "a super simple command",
}

func init() {
	flagsCmd.Flags().StringVarP(&strp, "string", "s", "foo", "a string")
	flagsCmd.Flags().BoolVarP(&boolp, "bool", "b", false, "a boolean")
	flagsCmd.Flags().IntVarP(&intp, "int", "i", 1, "an integer")

	rootCmd.AddCommand(flagsCmd)
}

func flagsFunc(cmd *cobra.Command, args []string) {
	fmt.Println("string:", strp)
	fmt.Println("boolean:", boolp)
	fmt.Println("integer:", intp)
}

func _() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
