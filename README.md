# Check TV (LG WebOS Controller)

A CLI tool written in Go to control LG TVs running WebOS using the SSAP protocol.

## Build

```bash
go build -o lg-webos-ssap main.go
```

## Usage

```bash
./lg-webos-ssap [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-addr` | `192.168.1.237:3000` | TV IP address and port |
| `-key-file` | `key` | Path to file containing the pairing key |
| `-cmd` | `(required)` | Command to execute (see below) |
| `-arg` | `""` | Argument for commands that require one |
| `-payload` | `""` | Optional JSON payload for `launch` command |
| `-use-socks5-proxy` | `""` | SOCKS5 proxy address (e.g. `127.0.0.1:1080`) |

### Commands

| Category | Command `-cmd` | Argument `-arg` | Description |
|----------|---------|----------|-------------|
| **Setup** | `initialize-key` | - | Initiates pairing and saves the key token. Run this first. |
| **System** | `info` | - | Get system information (model, version, etc). |
| | `toast` | Message | Show a toast notification on the TV. |
| **Apps** | `list-apps` | - | List all installed applications. |
| | `launch` | App ID/Name | Launch an application (e.g., "netflix", "youtube"). |
| | `close` | App ID/Name | Close a running application. |
| **Volume** | `vol-get` | - | Get current volume level. |
| | `vol-set` | `0-100` | Set volume to specific level. |
| | `vol-up` | - | Increase volume. |
| | `vol-down` | - | Decrease volume. |
| | `mute` | - | Mute audio. |
| | `un-mute` | - | Unmute audio. |
| **Channel** | `chan-get` | - | Get current channel info. |
| | `chan-up` | - | Channel up. |
| | `chan-down` | - | Channel down. |
| | `chan-set` | Channel ID | Switch to specific channel. |
| | `list-channels` | - | List available channels. |
| **Inputs** | `list-inputs` | - | List external inputs. |
| | `set-input` | Input ID | Switch input (e.g. HDMI_1). |
| **Media** | `play` | - | Play media. |
| | `pause` | - | Pause media. |
| | `stop` | - | Stop media. |
| | `rewind` | - | Rewind media. |
| | `fast-forward` | - | Fast forward media. |
| **Power** | `turn-off` | - | Turn off the TV. |

## Examples

1. **First time setup (Pairing with TV):**
   ```bash
   ./lg-webos-ssap -cmd initialize-key
   ```
   *Accept the prompt on your TV screen.*

2. **Control Volume:**
   ```bash
   ./lg-webos-ssap -cmd vol-get
   ./lg-webos-ssap -cmd vol-set -arg 15
   ./lg-webos-ssap -cmd mute
   ```

3. **Show Toast Message:**
   ```bash
   ./lg-webos-ssap -cmd toast -arg "Movie Night!"
   ```

4. **Launch Applications:**
   *You can use the exact App ID or the display name (case-insensitive).*
   ```bash
   ./lg-webos-ssap -cmd launch -arg "youtube"
   ./lg-webos-ssap -cmd launch -arg "com.webos.app.browser"
   ```

   *Launch with parameters (e.g. open a specific YouTube video):*
   ```bash
   ./lg-webos-ssap -cmd launch -arg youtube -payload '{"contentId":"v=dQw4w9WgXcQ"}'
   ```

5. **Specify a different TV address:**
   ```bash
   ./lg-webos-ssap -addr 192.168.1.50:3000 -cmd info
   ```

6. **Connect via SOCKS5 Proxy:**
   ```bash
   ./lg-webos-ssap -cmd info -use-socks5-proxy 127.0.0.1:1080
   ```
