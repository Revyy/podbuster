// keys: post keyboard events to one process (by PID) so an agent can drive
// the Podbuster window. Needs Accessibility permission for the host terminal.
//
//   keys trusted                      exit 0 when event posting is allowed
//   keys activate <pid>               bring the PID's app to the front
//   keys window <pid>                 print the PID's on-screen window id, exit 1 if none
//   keys tap <pid> <holdMs> <key>...  press keys together, hold, release
import AppKit
import CoreGraphics
import Foundation

let codes: [String: CGKeyCode] = [
    "a": 0x00, "s": 0x01, "d": 0x02, "w": 0x0D, "x": 0x07, "z": 0x06,
    "n": 0x2D, "p": 0x23, "r": 0x0F,
    "1": 0x12, "2": 0x13, "3": 0x14,
    "space": 0x31, "enter": 0x24, "escape": 0x35, "backspace": 0x33,
    "left": 0x7B, "right": 0x7C, "down": 0x7D, "up": 0x7E,
    "f12": 0x6F,
]

func fail(_ msg: String) -> Never {
    FileHandle.standardError.write((msg + "\n").data(using: .utf8)!)
    exit(2)
}

let args = CommandLine.arguments
guard args.count >= 2 else { fail("usage: keys trusted | activate <pid> | tap <pid> <holdMs> <key>...") }

switch args[1] {
case "trusted":
    exit(AXIsProcessTrusted() ? 0 : 1)
case "activate":
    guard args.count == 3, let pid = pid_t(args[2]),
          let app = NSRunningApplication(processIdentifier: pid) else { fail("activate: no app for pid") }
    app.activate()
    usleep(300_000)
case "window":
    guard args.count == 3, let pid = Int(args[2]) else { fail("usage: keys window <pid>") }
    let infos = CGWindowListCopyWindowInfo([.optionOnScreenOnly, .excludeDesktopElements], kCGNullWindowID) as? [[String: Any]] ?? []
    for w in infos where (w[kCGWindowOwnerPID as String] as? Int) == pid && (w[kCGWindowLayer as String] as? Int) == 0 {
        print(w[kCGWindowNumber as String] as? Int ?? 0)
        exit(0)
    }
    exit(1)
case "tap":
    guard args.count >= 5, let pid = pid_t(args[2]), let hold = UInt32(args[3]) else {
        fail("usage: keys tap <pid> <holdMs> <key>...")
    }
    let keys = args[4...].map { name -> CGKeyCode in
        guard let c = codes[name.lowercased()] else { fail("unknown key: \(name) (known: \(codes.keys.sorted().joined(separator: " ")))") }
        return c
    }
    let src = CGEventSource(stateID: .hidSystemState)
    for k in keys { CGEvent(keyboardEventSource: src, virtualKey: k, keyDown: true)!.postToPid(pid); usleep(5_000) }
    usleep(hold * 1000)
    for k in keys.reversed() { CGEvent(keyboardEventSource: src, virtualKey: k, keyDown: false)!.postToPid(pid); usleep(5_000) }
default:
    fail("unknown command: \(args[1])")
}
