golang program that connects to IP:PORT of a LG TV with webOs running SSAP protocol.
It should allow the user (via command-line args and flags) to call the most used SSAP capabilities, including: 

- Volume control: get and change
- Channel control: get and change up/down or to a specific channel
- Get TV info (version, model, etc)
- App control: list, launch or close an app (by name or id)
- Show toast message
- and all others possible commands the user might find relevant allowed by the SSAP protocol

Also add a special option `-cmd initialize-key` that will make a request to the tv with all possible the permissions and store the returned key in a local file for future use. This is intended to be run once to get the key, that latter can be used with any command without needing to re-authorize.

The `-cmd launch -arg youtube` should accept an additional optional argument to specify a string with json-payload to send to the app.

The default IP:PORT are 192.168.1.237:3000

There is a tv running webOS on that IP:PORT for testing.

Add an optional flag `-use-socks5-proxy 127.0.0.1:1080` that will make the program connect to the tv via a socks5 proxy



