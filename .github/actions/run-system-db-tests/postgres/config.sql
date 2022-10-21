ALTER SYSTEM SET ssl_ca_file TO '/tls-certs/ca-cert.pem';
ALTER SYSTEM SET ssl_cert_file TO '/tls-certs/server-cert.pem';
ALTER SYSTEM SET ssl_key_file TO '/tls-certs/server-key.pem';
ALTER SYSTEM SET ssl TO 'ON';
