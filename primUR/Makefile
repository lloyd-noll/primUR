# primUR Makefile
#
# Usage:
#   make install     – clone and build all dependencies
#   make install-bin – install primUR binary to ~/go/bin
#   make check       – verify all tools are in PATH
#   make             – build binary
#   make test        – run a test (set NEIDB and SCOPDB)
#   make clean       – remove build artifacts

BIN    = primUR
SRC    = primUR.go
PREFIX = $(HOME)/go/bin

EVOLBIO_REPOS = neighbors fur biobox prim

RUNTIME_DEPS = taxi neighbors land fur makeFurDb cleanSeq cres \
	midRoot nj plotTree fa2prim prim2tab scop \
	phylonium primer3_core datasets unzip

BUILD_DEPS = go

# --- Default target ------------------------------------------------------

all: $(BIN)

# --- Dependency checks ---------------------------------------------------

check-build:
	@echo "Checking build dependencies..."
	@for dep in $(BUILD_DEPS); do \
		command -v $$dep > /dev/null 2>&1 \
		&& echo "  [ok] $$dep" \
		|| { echo "  [missing] $$dep"; exit 1; }; \
	done

check-runtime:
	@echo "Checking runtime dependencies..."
	@for dep in $(RUNTIME_DEPS); do \
		command -v $$dep > /dev/null 2>&1 \
		&& echo "  [ok] $$dep" \
		|| echo "  [missing] $$dep  <- run 'make install'"; \
	done

check: check-build check-runtime

# --- Install dependencies ------------------------------------------------

install-evolbioinf:
	@echo "Installing evolbioinf packages..."
	@for repo in $(EVOLBIO_REPOS); do \
		echo "  cloning $$repo..."; \
		git clone https://github.com/evolbioinf/$$repo /tmp/$$repo 2>/dev/null || \
			git -C /tmp/$$repo pull; \
		echo "  building $$repo..."; \
		(cd /tmp/$$repo && bash setup.sh && make && \
		cp bin/* $(PREFIX)/ 2>/dev/null || true); \
		echo "  [ok] $$repo"; \
	done

install-phylonium:
	@command -v phylonium > /dev/null 2>&1 \
		&& echo "  [ok] phylonium already installed" \
		|| { \
		git clone https://github.com/evolbioinf/phylonium /tmp/phylonium 2>/dev/null || \
			git -C /tmp/phylonium pull; \
		make -C /tmp/phylonium; \
		cp /tmp/phylonium/phylonium $(PREFIX)/; \
		echo "  [ok] phylonium installed"; \
		}

install-primer3:
	@command -v primer3_core > /dev/null 2>&1 \
		&& echo "  [ok] primer3 already installed" \
		|| { \
		git clone https://github.com/primer3-org/primer3 /tmp/primer3 2>/dev/null || \
			git -C /tmp/primer3 pull; \
		make -C /tmp/primer3/src; \
		cp /tmp/primer3/src/primer3_core $(PREFIX)/; \
		echo "  [ok] primer3 installed"; \
		}

install-datasets:
	@command -v datasets > /dev/null 2>&1 \
		&& echo "  [ok] datasets already installed" \
		|| { \
		echo "  [manual] Install NCBI datasets CLI:"; \
		echo "           https://www.ncbi.nlm.nih.gov/datasets/docs/v2/download-and-install/"; \
		}

install: install-evolbioinf install-phylonium install-primer3 install-datasets
	@echo ""
	@echo "Done. Run 'make check' to verify all tools are available."

# --- Install binary ------------------------------------------------------

install-bin: $(BIN)
	mkdir -p $(PREFIX)
	cp $(BIN) $(PREFIX)/
	@echo "primUR installed to $(PREFIX)/primUR"
	@echo "Make sure $(PREFIX) is in your PATH."
	@echo "If not, add to ~/.zshrc or ~/.bashrc:"
	@echo "  export PATH=\$$PATH:$(PREFIX)"

# --- Build ---------------------------------------------------------------

$(BIN): $(SRC)
	go build -o $(BIN) $(SRC)

# --- Test ----------------------------------------------------------------

test: $(BIN)
	./$(BIN) -t query.txt -n $(NEIDB) -d $(SCOPDB) -f

# --- Clean ---------------------------------------------------------------

clean:
	rm -f $(BIN)
	rm -rf logs

.PHONY: all check check-build check-runtime \
	install install-evolbioinf install-phylonium install-primer3 install-datasets \
	install-bin test clean