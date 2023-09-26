
Simple AI Hub
=============


Hub of the AI node instances.  
Query nodes parallelly.  

```
                       --- node  
                    /  
simple-ai-chat --- hub --- node  
                    \  
                       --- node  
```

Setup
-----

1. Clone and rename folder to `hub_[hub_name]`.  
2. Put the simple-ai-node instance in folder `node_[node_name]` and run instances.  
3. Copy and create `node.csv` from `node.csv.example` to link to nodes.  
4. Copy and create `.env` from `.env.example`.


.env
----

* PORT  
The port of application.  

* HUB  
Hub name.  

* ID  
Hub ID, use a number.  
