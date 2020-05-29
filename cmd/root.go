// Package cmd root command
/*
Copyright © 2020 xiaowei <xw_cht.y@live.cn>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"
	"sync"

	"github.com/sharego/proxysocket/lib"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "proxysocket inbound outbound",
	Short: "Another socket proxy",
	Long:  `This proxy support tcp, udp and unix socket, like: tcp://127.0.0.1:80`,
	Run: func(cmd *cobra.Command, args []string) {
		apps := viper.AllSettings()
		if apps == nil || len(apps) == 0 {
			fmt.Fprintf(os.Stderr, "Empty Config file: %v\n", viper.ConfigFileUsed())
			return
		}
		var servers []*lib.ServerConfig
		for k := range apps {
			var s lib.ServerConfig
			if e := viper.UnmarshalKey(k, &s); e != nil {
				fmt.Fprintf(os.Stderr, "Parse Config file: %v, error: %s\n", viper.ConfigFileUsed(), e)
				return
			}
			s.Name = k
			fmt.Printf("key = %s, value=%v\n", k, s)
			servers = append(servers, &s)
		}

		wg := new(sync.WaitGroup)

		for _, s := range servers {
			pc := lib.ProxyChainTunnel{}
			wg.Add(1)
			go func(sc *lib.ServerConfig) {
				defer wg.Done()
				pc.Serve(sc)
				fmt.Printf("%s(%s) quit\n",sc.Name, sc.In)
			}(s)
		}

		wg.Wait()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.proxysocket)")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".proxysocket" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName(".proxysocket")
	}

	viper.SetConfigType("yaml")

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Fprintf(os.Stderr, "Parse config file: %s, failed: %s\n", viper.ConfigFileUsed(), err)
	}
}
