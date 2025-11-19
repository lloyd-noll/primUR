## Description
Automated candidate-free marker discovery as well as primer design and in silico
testing using species specific genomic regions.
## Introduction
Marker discovery often relies on prior knowledge of potential marker genes,
Simple Sequence Repeats (SSR) or Single Nucleotide Polymorphisms (SNPs). Species
specific genomic regions serve as ideal candidates for diagnostic markers. `dima`
combines a multitude of bioinformatic tools to automate the workflow of diagnostic
marker discovery from the discovery of unique genomic regions of the target
group or species through primer design to in silico testing.
## Make the Programs
Clone the repository and change to it.
```
git clone https://github.com/lloyd-noll/dima
cd dima
```
Install additional dependancies by running [setup.sh](scripts/setup.sh).
`bash scripts/setup.sh`

Make the programs.
`make`

The directory `bin` now contains the executeables. Add them to your `PATH`.