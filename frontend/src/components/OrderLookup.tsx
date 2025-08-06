import { useState } from "react";
import { Search, Package, AlertCircle, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { useToast } from "@/hooks/use-toast";

interface OrderData {
  order_uid: string;
  track_number: string;
  entry: string;
  delivery: {
    name: string;
    phone: string;
    zip: string;
    city: string;
    address: string;
    region: string;
    email: string;
  };
  payment: {
    transaction: string;
    request_id: string;
    currency: string;
    provider: string;
    amount: number;
    payment_dt: number;
    bank: string;
    delivery_cost: number;
    goods_total: number;
    custom_fee: number;
  };
  items: Array<{
    chrt_id: number;
    track_number: string;
    price: number;
    rid: string;
    name: string;
    sale: number;
    size: string;
    total_price: number;
    nm_id: number;
    brand: string;
    status: number;
  }>;
  locale: string;
  internal_signature: string;
  customer_id: string;
  delivery_service: string;
  shardkey: string;
  sm_id: number;
  date_created: string;
  oof_shard: string;
}

export const OrderLookup = () => {
  const [orderUid, setOrderUid] = useState("");
  const [orderData, setOrderData] = useState<OrderData | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const { toast } = useToast();

  const handleSearch = async () => {
    if (!orderUid.trim()) {
      toast({
        title: "Please enter an order UID",
        variant: "destructive",
      });
      return;
    }

    setLoading(true);
    setError("");
    setOrderData(null);

    try {
      const response = await fetch(`/order/${orderUid}`);
      
      if (response.status === 404) {
        const errorText = await response.text();
        setError(errorText || `Order ${orderUid} not found`);
        toast({
          title: "Order not found",
          description: `No order found with UID: ${orderUid}`,
          variant: "destructive",
        });
      } else if (response.ok) {
        const data = await response.json();
        setOrderData(data);
        toast({
          title: "Order found",
          description: `Successfully loaded order ${orderUid}`,
        });
      } else {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "Failed to fetch order";
      setError(errorMessage);
      toast({
        title: "Error",
        description: errorMessage,
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      handleSearch();
    }
  };

  const formatCurrency = (amount: number, currency: string) => {
    return new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: currency,
    }).format(amount / 100);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  return (
    <div className="space-y-6">
      {/* Search Section */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Package className="h-5 w-5" />
            Order Lookup
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-2">
            <Input
              placeholder="Enter order UID (e.g., b563feb7b2b84b6test)"
              value={orderUid}
              onChange={(e) => setOrderUid(e.target.value)}
              onKeyPress={handleKeyPress}
              className="flex-1"
            />
            <Button onClick={handleSearch} disabled={loading}>
              {loading ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <Search className="h-4 w-4" />
              )}
              {loading ? "Searching..." : "Search"}
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Error Display */}
      {error && (
        <Card className="border-destructive">
          <CardContent className="pt-6">
            <div className="flex items-center gap-2 text-destructive">
              <AlertCircle className="h-5 w-5" />
              <span className="font-medium">{error}</span>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Order Data Display */}
      {orderData && (
        <div className="space-y-4">
          {/* Order Summary */}
          <Card>
            <CardHeader>
              <CardTitle>Order Summary</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Order UID</p>
                  <p className="font-mono text-sm">{orderData.order_uid}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Track Number</p>
                  <p className="font-mono text-sm">{orderData.track_number}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Customer ID</p>
                  <p>{orderData.customer_id}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Date Created</p>
                  <p>{formatDate(orderData.date_created)}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Delivery Service</p>
                  <Badge variant="secondary">{orderData.delivery_service}</Badge>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Entry</p>
                  <Badge variant="outline">{orderData.entry}</Badge>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Delivery Information */}
          <Card>
            <CardHeader>
              <CardTitle>Delivery Information</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Name</p>
                  <p>{orderData.delivery.name}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Phone</p>
                  <p>{orderData.delivery.phone}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Email</p>
                  <p>{orderData.delivery.email}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Address</p>
                  <p>{orderData.delivery.address}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">City</p>
                  <p>{orderData.delivery.city}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Region</p>
                  <p>{orderData.delivery.region}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">ZIP</p>
                  <p>{orderData.delivery.zip}</p>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Payment Information */}
          <Card>
            <CardHeader>
              <CardTitle>Payment Information</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Transaction ID</p>
                  <p className="font-mono text-sm">{orderData.payment.transaction}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Provider</p>
                  <Badge variant="secondary">{orderData.payment.provider}</Badge>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Bank</p>
                  <p>{orderData.payment.bank}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Total Amount</p>
                  <p className="font-semibold text-lg">
                    {formatCurrency(orderData.payment.amount, orderData.payment.currency)}
                  </p>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Goods Total</p>
                  <p>{formatCurrency(orderData.payment.goods_total, orderData.payment.currency)}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Delivery Cost</p>
                  <p>{formatCurrency(orderData.payment.delivery_cost, orderData.payment.currency)}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Custom Fee</p>
                  <p>{formatCurrency(orderData.payment.custom_fee, orderData.payment.currency)}</p>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Items */}
          <Card>
            <CardHeader>
              <CardTitle>Order Items ({orderData.items.length})</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {orderData.items.map((item, index) => (
                  <div key={item.chrt_id}>
                    {index > 0 && <Separator />}
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 pt-4">
                      <div>
                        <p className="text-sm font-medium text-muted-foreground">Product</p>
                        <p className="font-medium">{item.name}</p>
                        <p className="text-sm text-muted-foreground">{item.brand}</p>
                      </div>
                      <div>
                        <p className="text-sm font-medium text-muted-foreground">Price</p>
                        <p>{formatCurrency(item.price, orderData.payment.currency)}</p>
                        {item.sale > 0 && (
                          <Badge variant="destructive" className="text-xs">
                            -{item.sale}%
                          </Badge>
                        )}
                      </div>
                      <div>
                        <p className="text-sm font-medium text-muted-foreground">Total Price</p>
                        <p className="font-semibold">
                          {formatCurrency(item.total_price, orderData.payment.currency)}
                        </p>
                      </div>
                      <div>
                        <p className="text-sm font-medium text-muted-foreground">Size</p>
                        <p>{item.size || "N/A"}</p>
                      </div>
                      <div>
                        <p className="text-sm font-medium text-muted-foreground">Status</p>
                        <Badge variant={item.status === 202 ? "success" : "secondary"}>
                          {item.status}
                        </Badge>
                      </div>
                      <div>
                        <p className="text-sm font-medium text-muted-foreground">SKU</p>
                        <p className="font-mono text-sm">{item.nm_id}</p>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
};