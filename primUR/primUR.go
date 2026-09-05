package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// config holds all CLI arguments and flags
type config struct {
	queryFile string
	neiDB     string
	scopDB    string
	verbose   bool
	checkOnly bool
	resume    bool
}

// requiredTools lists all external binaries the pipeline depends on
var requiredTools = []string{
	"taxi", "neighbors", "datasets", "unzip",
	"phylonium", "nj", "midRoot", "land", "plotTree",
	"makeFurDb", "fur", "cleanSeq", "cres", "fa2prim",
	"primer3_core", "prim2tab", "scop",
}

func main() {
	cfg := parseArgs()

	if err := checkDependencies(requiredTools); err != nil {
		log.Fatal(err)
	}

	var targets []string
	if _, err := os.Stat(cfg.queryFile); err == nil {
		// -t is a file
		targets, err = readLines(cfg.queryFile)
		if err != nil {
			log.Fatalf("Could not read query file: %v", err)
		}
	} else {
		// -t is a species name directly
		targets = []string{cfg.queryFile}
	}

	if err := setupLogDir(); err != nil {
		log.Fatalf("Could not create log directory: %v", err)
	}

	for _, target := range targets {
		if err := processTarget(target, cfg); err != nil {
			log.Printf("Error processing %q: %v", target, err)
		}
	}
}

func parseArgs() config {
	var cfg config

	flag.StringVar(&cfg.queryFile, "t", "", "Query file (one species per line), or a single species name [required]")
	flag.StringVar(&cfg.neiDB, "n", "", "Genome database path for taxi and neighbors [required]")
	flag.StringVar(&cfg.scopDB, "d", "", "NCBI nt/SCOP database path for scop [required]")
	flag.BoolVar(&cfg.verbose, "f", false, "Feedback to terminal (quiet by default)")
	flag.BoolVar(&cfg.checkOnly, "c", false, "Run up to (and including) tree plotting, then stop")
	flag.BoolVar(&cfg.resume, "r", false, "Resume after tree plotting (expects prior outputs)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage:
  primUR -t <query.txt> -n <neidb_path> -d <scop_db_path> [options]

Options:
`)
		flag.PrintDefaults()
	}

	flag.Parse()

	if cfg.checkOnly && cfg.resume {
		fmt.Fprintln(os.Stderr, "Error: choose either -c or -r, not both.")
		os.Exit(1)
	}

	if cfg.queryFile == "" || cfg.neiDB == "" || cfg.scopDB == "" {
		fmt.Fprintln(os.Stderr, "Error: -t, -n, and -d are all required.")
		flag.Usage()
		os.Exit(1)
	}

	return cfg
}

func checkDependencies(tools []string) error {
	var missing []string
	for _, tool := range tools {
		if _, err := exec.LookPath(tool); err != nil {
			missing = append(missing, tool)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing dependencies: %s", strings.Join(missing, ", "))
	}
	return nil
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines, scanner.Err()
}

func setupLogDir() error {
	err := os.MkdirAll("logs", 0755)
	if err != nil {
		return err
	}

	f, err := os.Create("logs/datasets.log")
	if err != nil {
		return err
	}
	defer f.Close()

	f, err = os.Create("logs/phylonium.log")
	if err != nil {
		return err
	}
	defer f.Close()

	f, err = os.Create("logs/fur.log")
	if err != nil {
		return err
	}
	defer f.Close()

	f, err = os.Create("logs/timer.log")
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}

func safeTarget(name string) string {
	return strings.ReplaceAll(name, " ", "_")
}

func logMsg(cfg config, msg string) {
	if cfg.verbose {
		fmt.Println(msg)
	}
}

func runCmd(stdin string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%s failed: %w\nOutput: %s", name, err, stderr.String())
	}
	return stdout.String(), nil
}

func processTarget(target string, cfg config) error {
	safe := safeTarget(target)
	logMsg(cfg, target+":")

	// --- Directory setup -------------------------------------------------
	// ustilago_hordei/
	// ├── taxids.txt
	// ├── fur/          taccs.txt, naccs.txt, URs.fasta, summary.txt, fur.db
	// ├── genomes/
	// │   ├── targets/  *.fna
	// │   └── neighbors/ *.fna
	// ├── primers/      primers.fasta, best_primers.fasta, scop.out
	// └── phylogeny/    tree.dist, tree.nwk, tree_clean.nwk

	for _, dir := range []string{
		safe + "/fur",
		safe + "/genomes/targets",
		safe + "/genomes/neighbors",
		safe + "/primers",
		safe + "/phylogeny",
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	if !cfg.resume {
		// --- Step 1: taxi + neighbors ------------------------------------
		taxiOut, err := runCmd("", "taxi", target, cfg.neiDB)
		if err != nil {
			return err
		}

		lines := strings.Split(taxiOut, "\n")
		fields := strings.Fields(lines[1])
		taxonID := fields[0]

		err = os.WriteFile(safe+"/taxids.txt", []byte(taxonID+"\n"), 0644)
		if err != nil {
			return err
		}

		neighOut, err := runCmd("", "neighbors", "-g", "-l", "-t", taxonID, cfg.neiDB)
		if err != nil {
			return err
		}

		var targets, neighbors []string
		for _, line := range strings.Split(neighOut, "\n") {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			switch fields[0] {
			case "t":
				targets = append(targets, fields[1])
			case "n":
				neighbors = append(neighbors, fields[1])
			}
		}

		err = os.WriteFile(safe+"/fur/taccs.txt", []byte(strings.Join(targets, "\n")+"\n"), 0644)
		if err != nil {
			return err
		}

		err = os.WriteFile(safe+"/fur/naccs.txt", []byte(strings.Join(neighbors, "\n")+"\n"), 0644)
		if err != nil {
			return err
		}

		logMsg(cfg, "  taxi and neighbors complete")

		// --- Step 2: download genomes ------------------------------------
		err = os.MkdirAll(safe+"/genomes/targets/tmp_unzip", 0755)
		if err != nil {
			return err
		}
		err = os.MkdirAll(safe+"/genomes/neighbors/tmp_unzip", 0755)
		if err != nil {
			return err
		}

		// TARGET LOOP: DOWNLOAD, UNZIP, MOVE
		for _, taccs := range targets {
			_, err := runCmd("", "datasets", "download", "genome", "accession", taccs, "--filename", safe+"/genomes/targets/"+taccs+".zip")
			if err != nil {
				return err
			}
			logMsg(cfg, "  Downloaded: "+taccs)
			_, err = runCmd("", "unzip", "-o", safe+"/genomes/targets/"+taccs+".zip", "-d", safe+"/genomes/targets/tmp_unzip")
			if err != nil {
				return err
			}
			// delete zip
			os.Remove(safe + "/genomes/targets/" + taccs + ".zip")

			filepath.Walk(safe+"/genomes/targets/tmp_unzip", func(path string, info os.FileInfo, err error) error {
				if filepath.Ext(path) == ".fna" {
					dest := safe + "/genomes/targets/t_" + filepath.Base(path)
					return os.Rename(path, dest)
				}
				return nil
			})
		}
		os.RemoveAll(safe + "/genomes/targets/tmp_unzip")

		// NEIGHBOR LOOP: DOWNLOAD, UNZIP, MOVE
		for _, naccs := range neighbors {
			_, err := runCmd("", "datasets", "download", "genome", "accession", naccs, "--filename", safe+"/genomes/neighbors/"+naccs+".zip")
			if err != nil {
				return err
			}
			logMsg(cfg, "  Downloaded: "+naccs)
			_, err = runCmd("", "unzip", "-o", safe+"/genomes/neighbors/"+naccs+".zip", "-d", safe+"/genomes/neighbors/tmp_unzip")
			if err != nil {
				return err
			}
			// delete zip
			os.Remove(safe + "/genomes/neighbors/" + naccs + ".zip")

			filepath.Walk(safe+"/genomes/neighbors/tmp_unzip", func(path string, info os.FileInfo, err error) error {
				if filepath.Ext(path) == ".fna" {
					dest := safe + "/genomes/neighbors/n_" + filepath.Base(path)
					return os.Rename(path, dest)
				}
				return nil
			})
		}
		os.RemoveAll(safe + "/genomes/neighbors/tmp_unzip")

		logMsg(cfg, "  All genome downloads complete")

		// --- Step 3: phylogenetic tree -----------------------------------
		targetFnas, err := filepath.Glob(safe + "/genomes/targets/*.fna")
		if err != nil {
			return err
		}
		neighborFnas, err := filepath.Glob(safe + "/genomes/neighbors/*.fna")
		if err != nil {
			return err
		}
		allFnas := append(targetFnas, neighborFnas...)

		distFile, err := os.Create(safe + "/phylogeny/tree.dist")
		if err != nil {
			return err
		}
		var phyStderr strings.Builder
		cmd := exec.Command("phylonium", allFnas...)
		cmd.Stdout = distFile
		cmd.Stderr = &phyStderr
		cmd.Run() // warnings are not errors
		distFile.Close()

		logMsg(cfg, "  phylonium complete")

		njOut, err := runCmd("", "nj", safe+"/phylogeny/tree.dist")
		if err != nil {
			return err
		}

		rootOut, err := runCmd(njOut, "midRoot")
		if err != nil {
			return err
		}

		landOut, err := runCmd(rootOut, "land")
		if err != nil {
			return err
		}

		err = os.WriteFile(safe+"/phylogeny/tree.nwk", []byte(landOut), 0644)
		if err != nil {
			return err
		}

		treeNwk, err := os.ReadFile(safe + "/phylogeny/tree.nwk")
		if err != nil {
			return err
		}
		step1 := strings.ReplaceAll(string(treeNwk), "'", "''")
		re := regexp.MustCompile(`(GC[AF]_[0-9]+\.[0-9]+)[^']*`)
		step2 := re.ReplaceAllString(step1, "$1")
		err = os.WriteFile(safe+"/phylogeny/tree_clean.nwk", []byte(step2), 0644)
		if err != nil {
			return err
		}

		_, err = runCmd("", "plotTree", safe+"/phylogeny/tree_clean.nwk")
		if err != nil {
			return err
		}
		logMsg(cfg, "  Tree construction complete")
	} else {
		logMsg(cfg, "  --resume set; starting at marker discovery for "+target)
	}

	if cfg.checkOnly {
		logMsg(cfg, "  --check set; stopping after tree plotting for "+target)
		return nil
	}

	// --- Step 4: FUR marker discovery ------------------------------------
	_, err := runCmd("", "makeFurDb", "-t", safe+"/genomes/targets", "-n", safe+"/genomes/neighbors", "-d", safe+"/fur/fur.db")
	if err != nil {
		return err
	}

	furOut, err := runCmd("", "fur", "-d", safe+"/fur/fur.db")
	if err != nil {
		return err
	}

	furFasta, err := runCmd(furOut, "cleanSeq")
	if err != nil {
		return err
	}

	cresOut, err := runCmd(furFasta, "cres")
	if err != nil {
		return err
	}

	err = os.WriteFile(safe+"/fur/URs.fasta", []byte(furFasta), 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(safe+"/fur/summary.txt", []byte(cresOut), 0644)
	if err != nil {
		return err
	}
	logMsg(cfg, "  fur analysis completed")

	// --- Step 5: primer design -------------------------------------------
	fa2primOut, err := runCmd("", "fa2prim", safe+"/fur/URs.fasta")
	if err != nil {
		return err
	}

	prim3Out, err := runCmd(fa2primOut, "primer3_core")
	if err != nil {
		return err
	}

	prim2tabOut, err := runCmd(prim3Out, "prim2tab")
	if err != nil {
		return err
	}

	// prim2tab format: # Penalty  Forward  Reverse  Internal
	type primerRow struct {
		penalty float64
		forward string
		reverse string
	}
	var rows []primerRow
	for _, line := range strings.Split(prim2tabOut, "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		var p float64
		fmt.Sscanf(fields[0], "%f", &p)
		rows = append(rows, primerRow{p, fields[1], fields[2]})
	}

	if len(rows) == 0 {
		logMsg(cfg, "  No primers found for "+target)
		return nil
	}

	// sort by penalty ascending (insertion sort)
	for i := 1; i < len(rows); i++ {
		for j := i; j > 0 && rows[j].penalty < rows[j-1].penalty; j-- {
			rows[j], rows[j-1] = rows[j-1], rows[j]
		}
	}

	// primers.fasta: all pairs with penalty <= 1
	var fastaLines []string
	for _, r := range rows {
		if r.penalty > 1.0 {
			break
		}
		fastaLines = append(fastaLines,
			fmt.Sprintf(">f penalty: %f", r.penalty),
			r.forward,
			">r",
			r.reverse,
		)
	}
	err = os.WriteFile(safe+"/primers/primers.fasta", []byte(strings.Join(fastaLines, "\n")+"\n"), 0644)
	if err != nil {
		return err
	}

	// best_primers.fasta: only the best pair
	best := rows[0]
	bestFasta := fmt.Sprintf(">f penalty: %f\n%s\n>r\n%s\n", best.penalty, best.forward, best.reverse)
	err = os.WriteFile(safe+"/primers/best_primers.fasta", []byte(bestFasta), 0644)
	if err != nil {
		return err
	}

	logMsg(cfg, "  primer3 complete")

	// --- Step 6: in silico specificity test ------------------------------
	scopOut, err := runCmd("", "scop", "-d", cfg.scopDB, "-t", safe+"/taxids.txt", safe+"/primers/best_primers.fasta")
	if err != nil {
		return err
	}
	err = os.WriteFile(safe+"/primers/scop.out", []byte(scopOut), 0644)
	if err != nil {
		return err
	}
	logMsg(cfg, "  scop complete")

	logMsg(cfg, "Finished run for "+safe+"\n")
	return nil
}