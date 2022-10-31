Flags  
- -chan=1234 -channel number, must be common for particopating peers (theres a default value if nil)
- -users=3 -> Mention number of participating users
- to add -> Private_key.pem share file 
- to add -> RSA key pair file


go run .\main.go .\p2p.go .\receiver.go .\interface.go .\mdns.go .\keygen.go .\acknowledger.go .\message.go .\flags.go


Features:
- mDNS self discovery between peers
- Easily defineable step, phase numbers

Changes:

- Phase structure -> interface - Done
- receiver - end acknowledgement to sender if phase received - Done
- Keygen wait for ack, and proceed - Done

To Add: 

- Time to die and resend