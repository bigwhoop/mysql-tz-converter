# MySQL Time Zone Converter

This CLI tool converts all fields of all datetime, date and timestamp columns in a MySQL database from
one time zone to another.
 
## Installation

    go install github.com/bigwhoop/mysql-tz-converter

## Instructions

    mysql-tz-converter -u user -p pass -h host -P port database from_tz to_tz
    mysql-tz-converter -u root -p secret some_db CET UTC
    mysql-tz-converter --help 

## Appendix

This tool uses MySQL's `CONVERT_TZ()` function. To use it with named time zones (like `UTC`) you should
make sure that the time zones are properly installed.

 * Windows: http://dev.mysql.com/downloads/timezones.html
 * Linux: `mysql_tzinfo_to_sql /usr/share/zoneinfo | mysql -u root mysql`