function r(o,t=""){const n=[];for(const e of o){const l=t?`${t} / ${e.name}`:e.name;n.push({label:l,value:e.id}),e.children&&e.children.length>0&&n.push(...r(e.children,l))}return n}export{r as f};
