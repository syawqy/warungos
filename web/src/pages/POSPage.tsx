import { useState } from 'react';
import MenuGrid from '../components/POS/MenuGrid';
import CartPanel from '../components/POS/CartPanel';
import PaymentModal from '../components/POS/PaymentModal';
import { useCart, useCreateOrder } from '../hooks/useOrders';
import type { OrderType } from '../types';

export default function POSPage() {
  const cart = useCart();
  const createOrder = useCreateOrder();
  const [showPayment, setShowPayment] = useState(false);

  const handlePay = async (method: string) => {
    try {
      await createOrder.mutateAsync({
        branch_id: 'a0000000-0000-0000-0000-000000000001', // TODO: get from user context
        customer_name: cart.customerName,
        order_type: cart.orderType,
        items: cart.items.map(item => ({
          menu_item_id: item.menu_item.id,
          quantity: item.quantity,
          unit_price: item.menu_item.price,
        })),
        discount: cart.discount,
      });
      cart.clearCart();
      setShowPayment(false);
    } catch (error) {
      console.error('Failed to create order:', error);
    }
  };

  return (
    <div className="flex h-full">
      {/* Menu Grid — left 60% */}
      <div className="w-3/5 h-full">
        <MenuGrid onAddToCart={cart.addItem} />
      </div>

      {/* Cart Panel — right 40% */}
      <div className="w-2/5 h-full">
        <CartPanel
          items={cart.items}
          customerName={cart.customerName}
          orderType={cart.orderType}
          subtotal={cart.subtotal}
          tax={cart.tax}
          discount={cart.discount}
          total={cart.total}
          itemCount={cart.itemCount}
          onRemoveItem={cart.removeItem}
          onUpdateQuantity={cart.updateQuantity}
          onSetCustomerName={cart.setCustomerName}
          onSetOrderType={cart.setOrderType}
          onSetDiscount={cart.setDiscount}
          onClearCart={cart.clearCart}
          onPay={() => setShowPayment(true)}
        />
      </div>

      {/* Payment Modal */}
      <PaymentModal
        isOpen={showPayment}
        total={cart.total}
        onPay={handlePay}
        onClose={() => setShowPayment(false)}
      />
    </div>
  );
}
