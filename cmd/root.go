package cmd

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// RootCmd 是无子命令调用时的基本命令.
var RootCmd = &cobra.Command{
	Use:   "fieldline",
	Short: "Fieldline is a tool for generating various field lines from discrete physical quantities.",
	Long: `Fieldline is a tool for generating various field lines from discrete physical quantities.

Usage:

        fieldline command [arguments]

The commands are:

        server      run web server of fieldline
        scalar      visualization from a scalar field
        vector      visualization from a vector field
        tensor      visualization from a tensor field

Use "fieldline help [command]" for more information about a command.

Additional help topics:

        streamline         description of streamline
        hyperstreamline    description of hyperstreamline
        contourline        description of contourline

Use "fieldline help [topic]" for more information about that topic.`,
	// 如果你的基本程序有一个与之关联的行为(action), 请去掉下行注释:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute 将所有的子命令添加到根命令, 并且设置其相关 flag.
// 该函数将由 main.main() 调用. 它只需调用 rootCmd 一次.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// 你将在这里定义程序 flag 和配置设置.
	// Cobra 支持支持全局的(persistent) flag,
	// 一旦定义这种 flag, 它将在你的程序中全局可用.
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.fieldline.yaml)")

	// Cobra 同时支持局部 flag, 它只在当前行为(action)被直接调用时运行.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig 读取配置文件和环境变量(必须已设置).
func initConfig() {
	if cfgFile != "" {
		// 使用来自 flag 的配置文件.
		viper.SetConfigFile(cfgFile)
	} else {
		// 查找 home 目录.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// 在 home 目录中搜索名为 ".fieldline" (无扩展名) 的配置文件.
		viper.AddConfigPath(home)
		viper.SetConfigName(".fieldline")
	}

	viper.AutomaticEnv() // 读取匹配的环境变量.

	// 如果找到了一个配置文件, 将其读入.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
