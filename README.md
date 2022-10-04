# stewardwriter

Write Steward messages to socket

## Functionality

1. Send a specified message continously at the interval specified
2. Start a watcher on a message folder, and when a new message is written to the folder it is forwarded to the specified socket. The message file will be deleted when the message is sent.

```bash
  -interval int
        the interval in seconds between sending messages (default 10)
  -messageFullPath string
        the full path to the message to send at intervals
  -socketFullPath string
        the full path to the steward socket file
  -watchFolder string
        the folder to watch for new messages to send
```
