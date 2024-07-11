# Talk2M GRE go

This project is a proof of concept example of GRE encapsulation compatible with Talk2M and eCatcher connectivity. 
eCatcher encapsulates broadcast traffic in a manner that differs from RFCs, and this project is only meant to work with eCatcher. 
This project uses google's [gopacket](https://github.com/google/gopacket) library for eBPFs support and packet inspection and requires libpcap. 

## Project status 
Talk2M usage is currently reserved for Ewon hardware and so this project is only for demonstration purposes and used to collect requirements for further development. 


## Running this code 
It is recommended to use [t2m-client-gre](https://github.com/it-hms/t2m-client-gre) which provides VPN client functionality.