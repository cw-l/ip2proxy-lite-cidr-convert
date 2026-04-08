# ip2proxyliteconvert

[![Go Report Card](https://goreportcard.com/badge/github.com/cw-l/ip2proxy-lite-cidr-convert)](https://goreportcard.com/report/github.com/cw-l/ip2proxy-lite-cidr-convert)
[![License: AGPL-3.0](https://img.shields.io/badge/License-AGPL_3.0-blue.svg)](https://opensource.org/licenses/AGPL-3.0)

**ip2proxyliteconvert** is a high-performance CLI utility written in Go for converting [IP2Proxy LITE](https://lite.ip2location.com/ip2proxy-lite) CSV database files into the MaxMind DB (MMDB) format. 

It is designed for cybersecurity researchers and sysadmins who need fast, local lookups of proxy, VPN, and data center IP ranges within their own applications or via standard MMDB readers.

MMDB enables ultra-fast, indexed IP lookups in applications and security pipelines, unlike flat CSV files which require linear scans.

## Key Features

* **Comprehensive Schema Support:** Supports all IP2Proxy LITE levels from **PX1 to PX12**.
* **Memory Efficient:** Processes large CSV files using streaming reads to avoid loading the entire dataset into memory.
* **Automated Filtering:** Automatically skips private, loopback, and reserved IP ranges (IPv4 and IPv6) to ensure database integrity.
* **CI/CD Ready:** Fully compatible with GoReleaser and includes a standard GNU-style Makefile.

---

## Quick Start

```bash
# 1. Download dataset. You will need to register for a free account at IP2Location.

# Replace {TOKEN} with your token and {DATABASE_CODE} with the database code, e.g., PX11LITECIDR, PX12LITECIDR
wget https://www.ip2location.com/download?token={TOKEN}&file={DATABASE_CODE}

unzip IP2PROXY-LITE-PX11.CIDR.ZIP

# 2. Convert to MMDB
ip2proxyliteconvert --in IP2PROXY-LITE-PX11.CIDR.CSV --db px11

# 3. Verify output
mmdblookup --file IP2PROXY-LITE-PX11.MMDB --ip 8.8.8.8
```

---

## Installation

### Method 1: Pre-built Binary (Recommended)
Download the latest `.deb`, `.rpm`, or compressed binary from the [Releases](https://github.com/cw-l/ip2proxy-lite-cidr-convert/releases) page.

**Note on Linux Paths:**
* Installing via `.deb` or `.rpm` (GoReleaser) places the binary in `/usr/bin/`.
* Manual builds via `make install` place the binary in `/usr/local/bin/`.
* *If both exist, `/usr/local/bin/` typically takes precedence in your `$PATH`.*

### Method 2: From Source (Requires Go 1.24+)
```bash
git clone [https://github.com/cw-l/ip2proxy-lite-cidr-convert.git](https://github.com/cw-l/ip2proxy-lite-cidr-convert.git)
cd ip2proxy-lite-cidr-convert
make
sudo make install
```

---

## Usage

The tool requires an input CSV file and the corresponding database level (px1-px12).
```bash
ip2proxyliteconvert --in IP2PROXY-LITE-PX11.CIDR.CSV --db px11
```
### Options
| Flag | Description |
| :--- | :--- |
| `--in` | Path to the source IP2Proxy LITE CSV file. |
| `--db` | The database level (e.g., `px1`, `px4`, `px11`). |
| `--out` | (Optional) Custom output filename (Default: `IP2PROXY-LITE-<DB>.MMDB`). |
| `-v` | Print version information. |

---

## Technical Details

### Schema Registry
The utility maps CSV columns based on the IP2Proxy LITE specifications. For example, a PX11 conversion includes:

* proxy_type, country_code, country_name, region_name, city_name, isp, domain, usage_type, asn, as, last_seen, threat, and provider.

### Performance
By utilizing github.com/maxmind/mmdbwriter, the tool produces highly compressed, search-optimized databases compatible with any MaxMind DB reader library (like maxminddb-golang or geoip2).

---

## Limitations & Caveats

* **Hardcoded Validation Threshold:** The current version (v{{.Version}}) will terminate if it encounters 100 consecutive malformed rows. This is a safety mechanism to prevent the generation of corrupt MMDB files if the incorrect `--db` level is selected. 
* **Planned Improvement:** Future releases will move toward a user-configurable threshold or a percentage-based error tolerance relative to the total record count.
* **Upstream Changes:** Significant changes to the IP2Proxy LITE CSV schema by the provider may trigger the validation threshold and require an update to this tool.
* **Validation Logic:** The converter includes a safety threshold. If the parser encounters more than 100 consecutive malformed rows (often caused by selecting the incorrect `--db` level for the input file), the process will terminate to prevent the generation of a corrupt MMDB.
* **Scope:** This is an independent community tool and is not affiliated with IP2Location or MaxMind.

---

## Development

### Makefile Targets
* `make build`: Compiles the binary to the root directory with version injection.
* `make tidy`: Syncs Go dependencies and cleans go.mod.
* `make install`: Installs the binary to /usr/local/bin.
* `make clean`: Removes binary and generated .MMDB files.

### Release Process
This project uses GoReleaser. To generate a local snapshot:
`goreleaser release --snapshot --clean`

---

## Credits
This utility uses the [IP2Location LITE database](https://lite.ip2location.com/ip2proxy-lite) for <a href="https://lite.ip2location.com">IP geolocation</a>. IP2Location is a registered trademark of Hexasoft Development Sdn Bhd. All other trademarks are the property of their respective owners.

Built with [mmdbwriter](https://github.com/maxmind/mmdbwriter) by [MaxMind](https://www.maxmind.com).