# 1. build bersih
rm -rf bin && ./scripts/build.sh

# 2. test hijau
go test ./...

# 3. vet + format
go vet ./...
gofmt -l .    # harus kosong

# 4. lint (kalau ada)
command -v golangci-lint >/dev/null && golangci-lint run ./... || echo "skip lint"

# 5. smoke test isolated
./test.sh

# 6. cek manager tidak import config
go list -deps ./cmd/pj | grep -q 'pj/internal/config' && echo "VIOLATION" || echo "ok"

# 7. cek tidak ada dead code
grep -r "internal/cli" --include='*.go' . || echo "cli is dead — remove"

# 8. tag
git log --oneline -10   # pastikan commit bersih, conventional commits