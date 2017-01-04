### checksum-test ###
checksum-test is a CLI tool that allows to detect silent network data corruption
in a TCP stream. When integrity mechanisms implemented at the different levels
of the ISO/OSI stack fail (e.g. Ethernet CRC32, IP header checksum, TCP checksum, etc.),
checksum-test performs checksum verification of the application level
payload and dumps on the filesystem all the messages that fail the consistency 
check. An example of a interesting use case is reported at
http://marcoguerri.github.io/linux/hardware/kernel/2016/06/19/mp30-data-corruption-part1.html

### Usage ###
Testing the tool requires a way to allow corrupted traffic to reach userspace
applications, which is normally not possible due to the various consistency
checks at several levels of the network stack, unless the corruption is applied
at the right moment. If, for instance, corruption of the application level payload 
happens at the queue discipline level when layer 4 consistency is offloaded to 
the NIC, it will be invisible to the TCP checksum, but it will be detected by the 
application level one.


### License ###
`checksum-test` is licensed under the GPLv3 license.
```
checksum-test - Tool which detects silent network data corruption

Copyright (C) 2016-2017 Marco Guerri <marco.guerri.dev@fastmail.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>


