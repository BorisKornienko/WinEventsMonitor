-- big thnx https://stackoverflow.com/questions/1385552/datatype-for-storing-ip-address-in-sql-server
-- https://stackoverflow.com/users/109122/rbarryyoung

-- create functions 

-- from string to binary

CREATE FUNCTION dbo.fnBinaryIPv4(@ip AS VARCHAR(15)) RETURNS BINARY(4)
AS
BEGIN
    DECLARE @bin AS BINARY(4)

    SELECT @bin = CAST( CAST( PARSENAME( @ip, 4 ) AS INTEGER) AS BINARY(1))
                + CAST( CAST( PARSENAME( @ip, 3 ) AS INTEGER) AS BINARY(1))
                + CAST( CAST( PARSENAME( @ip, 2 ) AS INTEGER) AS BINARY(1))
                + CAST( CAST( PARSENAME( @ip, 1 ) AS INTEGER) AS BINARY(1))

    RETURN @bin
END
go

-- and return from binary to string

CREATE FUNCTION dbo.fnDisplayIPv4(@ip AS BINARY(4)) RETURNS VARCHAR(15)
AS
BEGIN
    DECLARE @str AS VARCHAR(15) 

    SELECT @str = CAST( CAST( SUBSTRING( @ip, 1, 1) AS INTEGER) AS VARCHAR(3) ) + '.'
                + CAST( CAST( SUBSTRING( @ip, 2, 1) AS INTEGER) AS VARCHAR(3) ) + '.'
                + CAST( CAST( SUBSTRING( @ip, 3, 1) AS INTEGER) AS VARCHAR(3) ) + '.'
                + CAST( CAST( SUBSTRING( @ip, 4, 1) AS INTEGER) AS VARCHAR(3) );

    RETURN @str
END;
go

-- EXAMPLES

SELECT dbo.fnBinaryIPv4('192.65.68.201')
--should return 0xC04144C9
go

SELECT dbo.fnDisplayIPv4( 0xC04144C9 )
-- should return '192.65.68.201'
go



-- UPDATE:

-- I wanted to add that one way to address the inherent performance problems of scalar UDFs in SQL Server, but still retain the code-reuse of a function is to use an iTVF (inline table-valued function) instead. Here's how the first function above (string to binary) can be re-written as an iTVF:

CREATE FUNCTION dbo.itvfBinaryIPv4(@ip AS VARCHAR(15)) RETURNS TABLE
AS RETURN (
    SELECT CAST(
               CAST( CAST( PARSENAME( @ip, 4 ) AS INTEGER) AS BINARY(1))
            +  CAST( CAST( PARSENAME( @ip, 3 ) AS INTEGER) AS BINARY(1))
            +  CAST( CAST( PARSENAME( @ip, 2 ) AS INTEGER) AS BINARY(1))
            +  CAST( CAST( PARSENAME( @ip, 1 ) AS INTEGER) AS BINARY(1))
                AS BINARY(4)) As bin
        )
go

-- Here's it in the example:

SELECT bin FROM dbo.fnBinaryIPv4('192.65.68.201')
--should return 0xC04144C9
go

-- And here's how you would use it in an INSERT

-- INSERT INTo myIpTable
-- SELECT {other_column_values,...},
--        (SELECT bin FROM dbo.itvfBinaryIPv4('192.65.68.201'))