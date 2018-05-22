# Proxy check (pxcheck)

Requests through the proxy external IP, returns timings. First colon -
time to connect to proxy, second - overall query time, third -
difference, fifth - external IP. Time in nanoseconds. 

Usage example:

	pxcheck 1.2.3.4:8080
	130223411	192210200	61986789	1.2.3.4:8080	8.7.6.5	wgetip.com	

