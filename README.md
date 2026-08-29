# Blood-Donor-App
Blood Donor

## Runtime data

The application serves frontend files from `static/` and stores runtime data separately:

```text
data/
└── donors.json
uploads/
```

Uploaded images are stored in `/data/uploads/` and donor records are stored in `/data/donors.json`.

For a persistent disk mounted at `/data`, set:

```env
DATA_DIR=/data
```

The application then reads and writes donor records at `/data/donors.json` and uploaded images at `/data/uploads/`.

## Run

```bash
go run ./cmd/server
```