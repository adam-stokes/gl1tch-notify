import Foundation
import Network

// BUSDClient connects to the gl1tch event bus (Unix domain socket).
// It sends a registration frame, then streams newline-delimited JSON events.
// Publishing is supported via publish(event:payload:).
//
// Reconnects automatically with exponential backoff (1s → 30s).
final class BUSDClient: @unchecked Sendable {
    var onEvent: ((String, [String: Any]) -> Void)?
    var onConnectionChange: ((Bool) -> Void)?

    private(set) var isConnected = false

    private var connection: NWConnection?
    private let queue = DispatchQueue(label: "com.gl1tch.notify.busd", qos: .background)
    private var receiveBuffer = Data()
    private var backoff: TimeInterval = 1
    private let backoffMax: TimeInterval = 30
    private var stopped = false

    // MARK: — Start / Stop

    func start() {
        stopped = false
        connect()
    }

    func stop() {
        stopped = true
        connection?.cancel()
        connection = nil
    }

    // MARK: — Publish

    func publish(event: String, payload: [String: Any]) {
        guard isConnected, let conn = connection else { return }
        let frame: [String: Any] = ["action": "publish", "event": event, "payload": payload]
        guard let data = try? JSONSerialization.data(withJSONObject: frame),
              let line = String(data: data, encoding: .utf8) else { return }
        let wire = (line + "\n").data(using: .utf8)!
        conn.send(content: wire, completion: .idempotent)
    }

    // MARK: — Connect

    private func connect() {
        guard !stopped else { return }

        let path = socketPath()
        let endpoint = NWEndpoint.unix(path: path)
        // NWParameters() = no protocol stack — raw byte stream over the socket.
        let conn = NWConnection(to: endpoint, using: NWParameters())
        connection = conn

        conn.stateUpdateHandler = { [weak self] state in
            guard let self else { return }
            switch state {
            case .ready:
                self.backoff = 1
                self.isConnected = true
                DispatchQueue.main.async { self.onConnectionChange?(true) }
                self.sendRegistration(conn)
                self.receive(conn)
            case .failed, .cancelled:
                self.isConnected = false
                DispatchQueue.main.async { self.onConnectionChange?(false) }
                self.scheduleReconnect()
            case .waiting:
                // Socket not present yet — treat as disconnected.
                self.isConnected = false
                DispatchQueue.main.async { self.onConnectionChange?(false) }
                conn.cancel()
                self.scheduleReconnect()
            default:
                break
            }
        }
        conn.start(queue: queue)
    }

    private func sendRegistration(_ conn: NWConnection) {
        let reg = "{\"name\":\"glitch-notify\",\"subscribe\":[\"*\"]}\n"
        conn.send(content: reg.data(using: .utf8)!, completion: .idempotent)
    }

    private func receive(_ conn: NWConnection) {
        conn.receive(minimumIncompleteLength: 1, maximumLength: 65_536) { [weak self] content, _, isComplete, error in
            guard let self else { return }
            if let data = content {
                self.receiveBuffer.append(data)
                self.drainBuffer()
            }
            if error == nil && !isComplete {
                self.receive(conn)
            }
        }
    }

    private func drainBuffer() {
        let newline = UInt8(ascii: "\n")
        while let idx = receiveBuffer.firstIndex(of: newline) {
            let lineData = receiveBuffer[receiveBuffer.startIndex..<idx]
            receiveBuffer.removeSubrange(receiveBuffer.startIndex...idx)

            guard let json = try? JSONSerialization.jsonObject(with: lineData) as? [String: Any],
                  let event = json["event"] as? String,
                  let payload = json["payload"] as? [String: Any]
            else { continue }

            DispatchQueue.main.async {
                self.onEvent?(event, payload)
            }
        }
    }

    private func scheduleReconnect() {
        guard !stopped else { return }
        queue.asyncAfter(deadline: .now() + backoff) { [weak self] in
            guard let self, !self.stopped else { return }
            self.connect()
        }
        backoff = min(backoff * 2, backoffMax)
    }

    // MARK: — Socket path

    private func socketPath() -> String {
        if let dir = ProcessInfo.processInfo.environment["XDG_RUNTIME_DIR"] {
            return "\(dir)/glitch/bus.sock"
        }
        let cacheDir = FileManager.default.urls(for: .cachesDirectory, in: .userDomainMask).first!
        return cacheDir.appendingPathComponent("glitch/bus.sock").path
    }
}
