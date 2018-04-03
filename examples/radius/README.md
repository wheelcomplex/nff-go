To run tests for RADIUS lookup add the following to the configu of
your freeradius server:

File /etc/raddb/mods-config/files/authorize:
```
testuser1       Cleartext-Password := "testpass"
                BNG-ID := "4",
                BNG-Profile-ID := "18"

```
File /etc/raddb/dictionary:
```
$INCLUDE <PATH TO NFF-GO>/examples/radius/dictionary.bng
```
where <PATH TO NFF-GO> is value of `$NFF-GO` variable. After that run
radius server `radiusd -X` as root.
