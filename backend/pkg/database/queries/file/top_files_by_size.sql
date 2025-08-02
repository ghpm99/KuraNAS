select
    hf.id,
    hf.name,
    hf."size",
    hf."path"
from
    home_file hf
order by
    hf."size" desc
limit
    ?;