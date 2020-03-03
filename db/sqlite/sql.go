package sqlite

const getBlocksByMtimeQuery = `
select pos,data,mtime
from blocks b
where b.mtime > ?
order by b.mtime asc
limit ?
`

const FindBlocksInArea = `
select pos,data,mtime
from blocks b
where b.x >= ?1 and b.y >= ?2 and b.z >= ?3
  and b.x <= ?4 and b.y <= ?5 and b.z <= ?6
`

const countBlocksQuery = `
select count(*) from blocks b
`

const getBlockQuery = `
select pos,data,mtime from blocks b where b.pos = ?
`

const getTimestampQuery = `
select strftime('%s', 'now')
`
