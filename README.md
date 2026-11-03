# Rehearsal

Stress-test a small brand's monthly marketing plan before the money is spent. Go, stdlib only. See SPEC.md.

```
python data/synthetic.py            # synthetic demo brand
python reference/dump_fixtures.py   # golden fixtures from the Python reference engine
go test ./...                       # Go engine must match them
go run ./cmd/server                 # http://localhost:8080
```
