select
    hf.format,
    count(*) as total_itens,
    sum(hf."size") as total_size
from
    home_file hf
where
    hf."type" = ?
group by
    hf.format