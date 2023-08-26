---
title: Save draft mail in Zimbra web client using ChromeDP
date: 2020-07-03
description:
categories:
    - Programming
tags:
    - chromedp
    - golang
    - cobra
---
As an engineer, I want to automate everything as much as possible.
This CLI tool is created to save a draft mail in Zimbra web client.

Read config file:

```go
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

		// Search config in home directory with name ".zwc" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".zwc")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Using config file:", viper.ConfigFileUsed())
}
```

Create a new chromedp context:

```go
		allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Flag("headless", false))...)
		defer cancel()

		ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithDebugf(log.Printf))
		defer cancel()
```

Get flag values:

```go
		subject, err := cmd.Flags().GetString("subject")
		if err != nil {
			log.Fatal(err)
		}

		attachFile, err := cmd.Flags().GetString("attach")
		if err != nil {
			log.Fatal(err)
		}
```

And save draft mail:

```go
func zwcSaveDraft(subject, attachFile string) chromedp.Tasks {
	selName := `//input[@id="username"]`
	selPass := `//input[@id="password"]`

	return chromedp.Tasks{
		chromedp.Navigate(viper.GetString("url")),
		chromedp.WaitVisible(selPass),
		chromedp.SendKeys(selName, viper.GetString("username")),
		chromedp.SendKeys(selPass, viper.GetString("password")),
		chromedp.Submit(selPass),
		chromedp.WaitVisible(`//div[@id="z_userName"]`),
		chromedp.Click(`//div[@title="Compose"]`),
		chromedp.WaitVisible(`//div[@id="zb__App__tab_COMPOSE-1"]`),
		chromedp.SendKeys(`//input[@id="zv__COMPOSE-1_subject_control"]`, subject),
		chromedp.SendKeys(`//input[@type="file"]`, attachFile, chromedp.NodeVisible),
		chromedp.WaitVisible(`//a[@class="AttLink"]`),
		chromedp.Click(`//div[@id="zb__COMPOSE-1__SAVE_DRAFT"]`),
	}
}
```

GitHub repository: https://github.com/quantonganh/zwc