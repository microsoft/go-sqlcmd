set nocount on
declare @d1 datetime = '2022-03-05 14:01:02'
declare @d2 datetime2(4) = '2021-1-2 11:06:02.2'
declare @d3 datetimeoffset(6) = '2021-5-5'
declare @d4 smalldatetime = '2019-01-11 13:00:00'
declare @d5 time = '14:01:02'
declare @d6 date = '2011-02-03'

select @d1, @d2, @d3, @d4, @d5, @d6
