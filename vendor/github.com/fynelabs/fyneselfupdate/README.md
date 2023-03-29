# Fyne Selfupdate
An extension of the `selfupdate` repository that handles certain Fyne integrations.

## Simple setup for Fyne application

If you want to add self update support to your Fyne application, you can just just add the following code
```go
	// Used `selfupdatectl create-keys` followed by `selfupdatectl print-key`
	publicKey := ed25519.PublicKey{178, 103, 83, 57, 61, 138, 18, 249, 244, 80, 163, 162, 24, 251, 190, 241, 11, 168, 179, 41, 245, 27, 166, 70, 220, 254, 118, 169, 101, 26, 199, 129}

	// The public key above match the signature of the below file served by our CDN
	httpSource := selfupdate.NewHTTPSource(nil, "http://localhost/myprogram-{{.OS}}-{{.Arch}}{{.Ext}}")

	config := fyneselfupdate.NewConfigWithTimeout(a, w, time.Duration(1)*time.Minute,
		httpSource,
		selfupdate.Schedule{FetchOnStart: true, Interval: time.Hour * time.Duration(12)},
		publicKey)

	_, err := selfupdate.Manage(config)
	if err != nil {
		log.Println("Error while setting up update manager: ", err)
		return
	}
```

For a better understanding of how to use `selfupdatectl` to help you manage your self updating Fyne application deployment you can read the documentation [here](https://github.com/fynelabs/selfupdate/tree/main/cmd/selfupdatectl)

## Callbacks

You can also use following callbacks and directly create your custom selfupdate.Config structure to adapt presented graphical element to your exact need in your Fyne app.

```go
func NewUpgradeConfirmCallback(fyne.Window)

func NewUpgradeConfirmCallbackWithTimeout(fyne.Window, time.Duration)

func NewRestartConfirmCallback(win fyne.Window)

func NewRestartConfirmCallbackWithTimeout(win fyne.Window, timeout time.Duration)

func NewProgressCallback(win fyne.Window)

func NewExitCallback(app fyne.App, win fyne.Window)
```

## API Compatibility Promises
The main branch of `fyneselfupdate` is *not* guaranteed to have a stable API over time. Still we will try hard to not break its API unecessarily and will follow a proper versioning of our release. We will also keep it in sync and up to date with `fynelabs/selfupate`.

The `fyneselfupdate` package makes the following promises about API compatibility:
1. A list of all API-breaking changes will be documented in this README.
1. `fyneselfupdate` will strive for as few API-breaking changes as possible.

## License
BSD 3-Clause