# makeScopDb

**makeScopDb** is a helper script for [primUR](../primUR) that builds a taxonomy-linked BLAST nucleotide database for a target organism group. This is useful when using NCBI's full `nt` database for *in silico* specificity testing with `scop` would be impractical due to size or speed.

## Dependencies

| Tool | Source |
|------|--------|
| taxi, neighbors | github.com/evolbioinf/neighbors |
| sqlite3 | system package manager (e.g. `brew install sqlite3`) |
| datasets | github.com/ncbi/datasets |
| unzip | system package manager |
| awk, grep, tr | pre-installed on macOS/Linux |
| makeblastdb | NCBI BLAST+ (https://blast.ncbi.nlm.nih.gov/doc/blast-help/downloadblastdata.html) |

## Installation

Download the script and make it executable:

```bash
chmod +x makeScopDb
```

Optionally move it to a directory in your `PATH` so it can be called from anywhere:

```bash
mv makeScopDb ~/go/bin/
```

## Usage

```bash
makeScopDb "<target group>" [path/to/neidb]
```

The target group name must match a taxon recognized by `taxi`. The `neidb` path defaults to `~/DBs/neidb` if not provided.

**Example:**

```bash
makeScopDb "Ustilago hordei" ~/Data/DBs/neidb
```

This creates the following files in the current working directory:

```
ustilago_hordei.accessions.txt   – accession list
ustilago_hordei.accsIDs.tsv      – accession to TaxID mapping
ustilago_hordei.taxid_map.txt    – sequence ID to TaxID map
ustilago_hordei.fna              – concatenated FASTA
ustilago_hordeiDb.*              – BLAST database files
```

The resulting database can then be passed to `primUR` via the `-d` flag:

```bash
primUR -t query.txt -n neidb -d ustilago_hordeiDb
```

## Documentation

See `makeScopDb.pdf` for full documentation.
