import { OrderLookup } from "@/components/OrderLookup";

const Index = () => {
  return (
    <div className="min-h-screen bg-background">
      <div className="container mx-auto px-4 py-8">
        <div className="max-w-4xl mx-auto">
          <div className="text-center mb-8">
            <h1 className="text-4xl font-bold text-foreground mb-2">Order Management System</h1>
            <p className="text-lg text-muted-foreground">
              Search and view detailed order information by entering an order UID
            </p>
          </div>
          <OrderLookup />
        </div>
      </div>
    </div>
  );
};

export default Index;
