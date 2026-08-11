terminal = go mod init whatsmeow-test

main.go = file add

**1. Create a new folder outside the whatsmeow repo** (don't put it inside the cloned repo)
```powershell
go mod init whatsmeow-test
```

**2. Create `main.go`**

**3. Install dependencies**
```powershell
go mod tidy
```

**4. Run it**
```powershell
go run main.go
```