### data-integrity-check ###
data-integrity-check is a CLI tool that allows to detect silent network data corruption
in a TCP stream. When integrity mechanisms implemented at the different levels
of the ISO/OSI stack fail (e.g. Ethernet CRC32, IP header checksum, TCP checksum, etc.),
data-integrity-check performs checksum verification of the application level
payload and dumps on the filesystem all the messages that fail the consistency check.
