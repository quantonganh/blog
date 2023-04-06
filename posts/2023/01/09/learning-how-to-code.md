---
title: Learning how to code
date: 2023-01-09
description:
category:
- Programming
tags:
- golang
- chromedp
- powershell
---
If you find yourself doing the same task repeatedly, consider learning how to code.

For instance, a friend of mine had to perform the following tasks on daily basis:

- logging into a internal portal
- selecting certain menus
- exporting some `.xlsx` reports
- opening various excel files
- clicking on the "Refresh All" button to update external data
- sending an email

Here's the code that I assisted him with: https://github.com/quantonganh/ims

The initial steps can be done by using [chromedp](https://github.com/chromedp/chromedp).
To help my friend overcome the first obstacle, I explored various Golang libraries, such as:

- the ones designed for working with `.xlsx`: https://github.com/qax-os/excelize
- or calling COM: https://github.com/go-ole/go-ole

Unfortunately, none of these libraries provided the effective in addressing the issue.

Finally, I discovered that using a [powershell script](https://github.com/quantonganh/ims/blob/main/cmd/refresh_all.ps1) to refresh the data:

```powershell
$Excel=New-Object -ComObject Excel.Application
$Excel.Visible = $true

for ($i=0; $i -lt $($args.Length - 1); $i++)
{
    $inputWb = $Excel.Workbooks.Open($($args[$i]))
}

$outputWb=$Excel.Workbooks.Open($($args[$($args.Length - 1)]))
$outputWb.RefreshAll()
While (($outputWb.Sheets | ForEach-Object {$_.QueryTables | ForEach-Object {if($_.QueryTable.Refreshing){$true}}}))
{
    Start-Sleep -Seconds 1
}
$outputWb.Save()
$outputWb.Close()
$Excel.Quit()
```

as well as [calling this from Golang](https://github.com/quantonganh/ims/blob/f10be5ea49ed48e0ce77e371cc1c8e3c68fbbd04/cmd/formula.go#L245) was an effective solution.

The second challenge involved the inability to use the internal Wi-Fi network to send emails due to ISP blocking port 587. 
In order to address this issue, I utilized the `netsh` command to [switch to a different Wi-Fi network](https://github.com/quantonganh/ims/blob/f10be5ea49ed48e0ce77e371cc1c8e3c68fbbd04/cmd/formula.go#L120) 
that could be used to send the email, and then switch back.

Additionally, I employed a technique to ensure a connection to the SMTP port prior to sending the email:

```go
		retry(time.Second, 30*time.Second, func() bool {
			address := net.JoinHostPort(conf.SMTP.Host, strconv.Itoa(conf.SMTP.Port))
			log.Printf("connecting to the %s", address)
			_, err := net.Dial("tcp", address)
			if err == nil {
				return true
			}
			return false
		})
```

```go
func retry(interval, timeout time.Duration, f func() bool) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	to := time.NewTimer(timeout)
	defer to.Stop()
	for {
		select {
		case <-ticker.C:
			if f() {
				return
			}
		case <-to.C:
			log.Info().Msgf("timed out after %s", timeout.String())
			return
		}
	}
}
```