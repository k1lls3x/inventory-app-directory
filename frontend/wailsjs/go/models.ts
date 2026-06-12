export namespace model {
	
	export class DailyStock {
	    date: string;
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new DailyStock(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.total = source["total"];
	    }
	}
	export class DashboardData {
	    total_stock: number;
	    item_count: number;
	    monthly_orders: number;
	    new_items: number;
	
	    static createFrom(source: any = {}) {
	        return new DashboardData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total_stock = source["total_stock"];
	        this.item_count = source["item_count"];
	        this.monthly_orders = source["monthly_orders"];
	        this.new_items = source["new_items"];
	    }
	}
	export class Inbound {
	    inbound_id: number;
	    item_id: number;
	    supplier_id: number;
	    quantity: number;
	    // Go type: time
	    received_at: any;
	    received_by?: number;
	    document_no?: string;
	    note?: string;
	    warehouse_id: number;
	
	    static createFrom(source: any = {}) {
	        return new Inbound(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.inbound_id = source["inbound_id"];
	        this.item_id = source["item_id"];
	        this.supplier_id = source["supplier_id"];
	        this.quantity = source["quantity"];
	        this.received_at = this.convertValues(source["received_at"], null);
	        this.received_by = source["received_by"];
	        this.document_no = source["document_no"];
	        this.note = source["note"];
	        this.warehouse_id = source["warehouse_id"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class InboundDetails {
	    inbound_id: number;
	    item_id: number;
	    supplier_id: number;
	    warehouse_id: number;
	    received_by?: number;
	    date: string;
	    received_at: string;
	    name: string;
	    sku: string;
	    supplier: string;
	    quantity: number;
	    warehouse: string;
	    receiver_name: string;
	    document_no: string;
	    note: string;
	
	    static createFrom(source: any = {}) {
	        return new InboundDetails(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.inbound_id = source["inbound_id"];
	        this.item_id = source["item_id"];
	        this.supplier_id = source["supplier_id"];
	        this.warehouse_id = source["warehouse_id"];
	        this.received_by = source["received_by"];
	        this.date = source["date"];
	        this.received_at = source["received_at"];
	        this.name = source["name"];
	        this.sku = source["sku"];
	        this.supplier = source["supplier"];
	        this.quantity = source["quantity"];
	        this.warehouse = source["warehouse"];
	        this.receiver_name = source["receiver_name"];
	        this.document_no = source["document_no"];
	        this.note = source["note"];
	    }
	}
	export class JSONNullFloat64 {
	    Float64: number;
	    Valid: boolean;
	
	    static createFrom(source: any = {}) {
	        return new JSONNullFloat64(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Float64 = source["Float64"];
	        this.Valid = source["Valid"];
	    }
	}
	export class Item {
	    item_id: number;
	    sku: string;
	    name: string;
	    description?: string;
	    uom: string;
	    reorder_level: number;
	    reorder_qty: number;
	    price: JSONNullFloat64;
	    cost: JSONNullFloat64;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Item(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.item_id = source["item_id"];
	        this.sku = source["sku"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.uom = source["uom"];
	        this.reorder_level = source["reorder_level"];
	        this.reorder_qty = source["reorder_qty"];
	        this.price = this.convertValues(source["price"], JSONNullFloat64);
	        this.cost = this.convertValues(source["cost"], JSONNullFloat64);
	        this.created_at = this.convertValues(source["created_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ItemBrief {
	    item_id: number;
	    name: string;
	    sku: string;
	
	    static createFrom(source: any = {}) {
	        return new ItemBrief(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.item_id = source["item_id"];
	        this.name = source["name"];
	        this.sku = source["sku"];
	    }
	}
	export class ItemFilter {
	    SKU?: string;
	    Name?: string;
	    UOM?: string;
	    PriceMin?: number;
	    PriceMax?: number;
	
	    static createFrom(source: any = {}) {
	        return new ItemFilter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.SKU = source["SKU"];
	        this.Name = source["Name"];
	        this.UOM = source["UOM"];
	        this.PriceMin = source["PriceMin"];
	        this.PriceMax = source["PriceMax"];
	    }
	}
	export class ItemTurnoverByWarehouse {
	    warehouse: string;
	    name: string;
	    sku: string;
	    received: number;
	    shipped: number;
	
	    static createFrom(source: any = {}) {
	        return new ItemTurnoverByWarehouse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.warehouse = source["warehouse"];
	        this.name = source["name"];
	        this.sku = source["sku"];
	        this.received = source["received"];
	        this.shipped = source["shipped"];
	    }
	}
	export class ItemWithStock {
	    stock_id: number;
	    item_id: number;
	    warehouse_id: number;
	    name: string;
	    sku: string;
	    warehouse: string;
	    quantity: number;
	
	    static createFrom(source: any = {}) {
	        return new ItemWithStock(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.stock_id = source["stock_id"];
	        this.item_id = source["item_id"];
	        this.warehouse_id = source["warehouse_id"];
	        this.name = source["name"];
	        this.sku = source["sku"];
	        this.warehouse = source["warehouse"];
	        this.quantity = source["quantity"];
	    }
	}
	
	export class Movement {
	    type: string;
	    movement_id: number;
	    item_id: number;
	    item_name: string;
	    quantity: number;
	    // Go type: time
	    date: any;
	    warehouse_id: number;
	    warehouse_name: string;
	    supplier_id?: number;
	    supplier_name?: string;
	    destination?: string;
	    // Go type: time
	    shipped_at?: any;
	
	    static createFrom(source: any = {}) {
	        return new Movement(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.movement_id = source["movement_id"];
	        this.item_id = source["item_id"];
	        this.item_name = source["item_name"];
	        this.quantity = source["quantity"];
	        this.date = this.convertValues(source["date"], null);
	        this.warehouse_id = source["warehouse_id"];
	        this.warehouse_name = source["warehouse_name"];
	        this.supplier_id = source["supplier_id"];
	        this.supplier_name = source["supplier_name"];
	        this.destination = source["destination"];
	        this.shipped_at = this.convertValues(source["shipped_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class OutboundDetails {
	    outbound_id: number;
	    item_id: number;
	    warehouse_id: number;
	    shipped_by?: number;
	    date: string;
	    shipped_at: string;
	    name: string;
	    sku: string;
	    destination: string;
	    quantity: number;
	    warehouse: string;
	    shipper_name: string;
	    document_no: string;
	    note: string;
	
	    static createFrom(source: any = {}) {
	        return new OutboundDetails(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.outbound_id = source["outbound_id"];
	        this.item_id = source["item_id"];
	        this.warehouse_id = source["warehouse_id"];
	        this.shipped_by = source["shipped_by"];
	        this.date = source["date"];
	        this.shipped_at = source["shipped_at"];
	        this.name = source["name"];
	        this.sku = source["sku"];
	        this.destination = source["destination"];
	        this.quantity = source["quantity"];
	        this.warehouse = source["warehouse"];
	        this.shipper_name = source["shipper_name"];
	        this.document_no = source["document_no"];
	        this.note = source["note"];
	    }
	}
	export class Stock {
	    stock_id: number;
	    item_id: number;
	    quantity: number;
	    warehouse_id: number;
	    // Go type: time
	    last_updated: any;
	
	    static createFrom(source: any = {}) {
	        return new Stock(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.stock_id = source["stock_id"];
	        this.item_id = source["item_id"];
	        this.quantity = source["quantity"];
	        this.warehouse_id = source["warehouse_id"];
	        this.last_updated = this.convertValues(source["last_updated"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class StockTransferDetails {
	    transfer_id: number;
	    item_id: number;
	    from_warehouse_id: number;
	    to_warehouse_id: number;
	    date: string;
	    item_name: string;
	    sku: string;
	    from_warehouse: string;
	    to_warehouse: string;
	    quantity: number;
	    note: string;
	
	    static createFrom(source: any = {}) {
	        return new StockTransferDetails(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.transfer_id = source["transfer_id"];
	        this.item_id = source["item_id"];
	        this.from_warehouse_id = source["from_warehouse_id"];
	        this.to_warehouse_id = source["to_warehouse_id"];
	        this.date = source["date"];
	        this.item_name = source["item_name"];
	        this.sku = source["sku"];
	        this.from_warehouse = source["from_warehouse"];
	        this.to_warehouse = source["to_warehouse"];
	        this.quantity = source["quantity"];
	        this.note = source["note"];
	    }
	}
	export class Supplier {
	    supplier_id: number;
	    name: string;
	    inn?: string;
	    contact_person?: string;
	    phone?: string;
	    email?: string;
	
	    static createFrom(source: any = {}) {
	        return new Supplier(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.supplier_id = source["supplier_id"];
	        this.name = source["name"];
	        this.inn = source["inn"];
	        this.contact_person = source["contact_person"];
	        this.phone = source["phone"];
	        this.email = source["email"];
	    }
	}
	export class User {
	    user_id: number;
	    username: string;
	    full_name: string;
	    role: string;
	
	    static createFrom(source: any = {}) {
	        return new User(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.user_id = source["user_id"];
	        this.username = source["username"];
	        this.full_name = source["full_name"];
	        this.role = source["role"];
	    }
	}
	export class UserUpdate {
	    user_id: number;
	    username: string;
	    full_name: string;
	    role: string;
	    password?: string;
	
	    static createFrom(source: any = {}) {
	        return new UserUpdate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.user_id = source["user_id"];
	        this.username = source["username"];
	        this.full_name = source["full_name"];
	        this.role = source["role"];
	        this.password = source["password"];
	    }
	}
	export class Warehouse {
	    warehouse_id: number;
	    name: string;
	    location?: string;
	
	    static createFrom(source: any = {}) {
	        return new Warehouse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.warehouse_id = source["warehouse_id"];
	        this.name = source["name"];
	        this.location = source["location"];
	    }
	}

}

