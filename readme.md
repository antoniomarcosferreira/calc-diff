# This tool run a list of calc_test in two differents runtimes and comppar results.

# To compile 
go build compar.go 

# to run (Linux)
./compar tests.csv https://rt-prod.taxweb.com.br:443/taxgateway/webapi/taxrules/calctaxes https://rt-extrafarma.taxweb.com.br:443/taxgateway/webapi/taxrules/calctaxes
