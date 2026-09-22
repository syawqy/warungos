import { useState } from 'react';
import { X, CreditCard, Smartphone, Banknote, CheckCircle } from 'lucide-react';
import { formatCurrency } from '../../lib/utils';

interface PaymentModalProps {
  isOpen: boolean;
  total: number;
  onPay: (method: string) => void;
  onClose: () => void;
}

const PAYMENT_METHODS = [
  { id: 'cash', label: 'Tunai', icon: Banknote, color: 'bg-green-100 text-green-700' },
  { id: 'qris', label: 'QRIS', icon: Smartphone, color: 'bg-blue-100 text-blue-700' },
  { id: 'card', label: 'Kartu', icon: CreditCard, color: 'bg-purple-100 text-purple-700' },
  { id: 'ewallet', label: 'E-Wallet', icon: Smartphone, color: 'bg-orange-100 text-orange-700' },
];

export default function PaymentModal({ isOpen, total, onPay, onClose }: PaymentModalProps) {
  const [selected, setSelected] = useState('cash');
  const [paid, setPaid] = useState(false);
  const [cashAmount, setCashAmount] = useState('');

  if (!isOpen) return null;

  const change = cashAmount ? Math.max(0, Number(cashAmount) - total) : 0;

  const handlePay = () => {
    if (selected === 'cash' && Number(cashAmount) < total) return;
    setPaid(true);
    setTimeout(() => {
      onPay(selected);
      setPaid(false);
      setCashAmount('');
    }, 2000);
  };

  if (paid) {
    return (
      <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
        <div className="bg-white rounded-2xl p-8 max-w-sm w-full text-center">
          <CheckCircle className="w-16 h-16 text-green-500 mx-auto mb-4" />
          <h3 className="text-xl font-bold mb-2">Pembayaran Berhasil!</h3>
          <p className="text-gray-500">Pesanan sedang diproses...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white rounded-2xl max-w-md w-full mx-4 overflow-hidden">
        {/* Header */}
        <div className="p-4 border-b flex items-center justify-between">
          <h3 className="font-bold text-lg">Pembayaran</h3>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Total */}
        <div className="p-6 text-center bg-gray-50">
          <p className="text-sm text-gray-500">Total yang harus dibayar</p>
          <p className="text-3xl font-bold text-warung-orange">{formatCurrency(total)}</p>
        </div>

        {/* Payment Methods */}
        <div className="p-4">
          <p className="text-sm font-medium text-gray-700 mb-3">Metode Pembayaran</p>
          <div className="grid grid-cols-2 gap-2">
            {PAYMENT_METHODS.map(method => {
              const Icon = method.icon;
              return (
                <button
                  key={method.id}
                  onClick={() => setSelected(method.id)}
                  className={`p-3 rounded-xl border-2 flex items-center gap-2 transition-all ${
                    selected === method.id
                      ? 'border-warung-orange bg-orange-50'
                      : 'border-gray-200 hover:border-gray-300'
                  }`}
                >
                  <div className={`p-1.5 rounded-lg ${method.color}`}>
                    <Icon className="w-4 h-4" />
                  </div>
                  <span className="text-sm font-medium">{method.label}</span>
                </button>
              );
            })}
          </div>
        </div>

        {/* Cash Input */}
        {selected === 'cash' && (
          <div className="px-4 pb-4">
            <label className="text-sm text-gray-500 mb-1 block">Jumlah Uang</label>
            <input
              type="number"
              value={cashAmount}
              onChange={(e) => setCashAmount(e.target.value)}
              placeholder="Masukkan jumlah uang"
              className="input-field"
            />
            {Number(cashAmount) >= total && (
              <p className="text-sm text-green-600 mt-1">
                Kembalian: {formatCurrency(change)}
              </p>
            )}
          </div>
        )}

        {/* Pay Button */}
        <div className="p-4 border-t">
          <button
            onClick={handlePay}
            disabled={selected === 'cash' && (!cashAmount || Number(cashAmount) < total)}
            className="w-full btn-primary py-3 text-lg"
          >
            Bayar {formatCurrency(total)}
          </button>
        </div>
      </div>
    </div>
  );
}
