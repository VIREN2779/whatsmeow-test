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

**4. Install a C compiler**
1. **MinGW-w64** via https://www.msys2.org/ (install msys2, then run `pacman -S mingw-w64-ucrt-x86_64-gcc` inside it)
2. Add its `bin` folder to PATH
3. Enable CGO:
   ```powershell
   $env:CGO_ENABLED=1
   go run main.go
   ```