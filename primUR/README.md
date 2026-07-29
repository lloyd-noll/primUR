# primUR

**primUR** is a pipeline for the automated discovery of taxon-specific diagnostic PCR primers. Given a list of target organisms, it identifies unique genomic regions (URs) absent from related neighbor genomes, designs primers against those regions, and tests their *in silico* specificity.

## Dependencies

primUR requires the following tools to be installed:

| Package | Programs | Source |
|---------|----------|--------|
| neighbors | taxi, neighbors, land | github.com/evolbioinf/neighbors |
| fur | fur, makeFurDb, cleanSeq | github.com/evolbioinf/fur |
| biobox | cres, midRoot, nj, plotTree | github.com/evolbioinf/biobox |
| prim | fa2prim, prim2tab, scop | github.com/evolbioinf/prim |
| phylonium | phylonium | github.com/evolbioinf/phylonium |
| primer3 | primer3_core | github.com/primer3-org/primer3 |
| datasets | datasets | github.com/ncbi/datasets |

Go (1.18 or later) is required to build primUR.

## Installation

Clone the repository:
```bash
git clone https://github.com/lloyd-noll/primUR
cd primUR
```

Install all pipeline dependencies:
```bash
make install
```

Build the binary:
```bash
make
```

Install `primUR` to `~/go/bin`:
```bash
make install-bin
```

Make sure `~/go/bin` is in your `PATH`. If not, add this to your `~/.zshrc` or `~/.bashrc`:
```bash
export PATH=$PATH:~/go/bin
```

Verify all dependencies are available:
```bash
make check
```

## Usage

Create a query file with one target species per line:
```
Ustilago hordei
```

Run the pipeline:
```bash
primUR -t query.txt -n <neidb_path> -d <scopdb_path>
```

### Options

| Flag | Description |
|------|-------------|
| `-t` | Query file (one target species per line) |
| `-n` | Genome database path for taxi and neighbors |
| `-d` | BLAST database path for scop |
| `-f` | Verbose output |
| `-c` | Stop after tree construction |
| `-r` | Resume after tree construction |

### Output

For each target species, primUR creates a directory with the following structure:
```
<target>/
├── taxids.txt
├── fur/
│   ├── taccs.txt
│   ├── naccs.txt
│   ├── URs.fasta
│   ├── summary.txt
│   └── fur.db
├── genomes/
│   ├── targets/
│   └── neighbors/
├── primers/
│   ├── primers.fasta
│   ├── best_primers.fasta
│   └── scop.out
└── phylogeny/
    ├── tree.dist
    ├── tree.nwk
    └── tree_clean.nwk
```

## Documentation

See `primUR.pdf` for full documentation.

## Author

Lloyd-Eddie Noll - Master's student in Biology, Universität Hamburg  
Thesis work at the Max Planck Institute for Evolutionary Biology, Plön
