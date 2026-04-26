import { Link } from 'react-router-dom';
import { ArrowRight, CheckCircle, XCircle, AlertCircle, Info, X } from 'lucide-react';
import { useState, useEffect, createContext, useContext, useCallback } from 'react';

const ToastContext = createContext();

export function ToastProvider({ children }) {
  const [toasts, setToasts] = useState([]);

  const addToast = useCallback((toast) => {
    const id = Date.now();
    setToasts((prev) => [...prev, { ...toast, id }]);
    setTimeout(() => {
      setToasts((prev) => prev.filter((t) => t.id !== id));
    }, 5000);
  }, []);

  const removeToast = useCallback((id) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  return (
    <ToastContext.Provider value={{ addToast }}>
      {children}
      <ToastContainer toasts={toasts} removeToast={removeToast} />
    </ToastContext.Provider>
  );
}

export function useToast() {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error('useToast must be used within ToastProvider');
  }
  return context;
}

function ToastContainer({ toasts, removeToast }) {
  return (
    <div className="toast-container">
      {toasts.map((toast) => (
        <Toast key={toast.id} toast={toast} onClose={() => removeToast(toast.id)} />
      ))}
    </div>
  );
}

function Toast({ toast, onClose }) {
  const icons = {
    success: <CheckCircle size={20} />,
    error: <XCircle size={20} />,
    warning: <AlertCircle size={20} />,
    info: <Info size={20} />,
  };

  return (
    <div className={`toast toast-${toast.type}`}>
      <div className="toast-icon">{icons[toast.type]}</div>
      <div className="toast-content">
        {toast.title && <div className="toast-title">{toast.title}</div>}
        <div className="toast-message">{toast.message}</div>
      </div>
      <button className="toast-close" onClick={onClose}>
        <X size={14} />
      </button>
      <div className="toast-progress" />
    </div>
  );
}

export function StatCard({ label, value, icon: Icon, href, change, delay = 0 }) {
  const content = (
    <div 
      className="group bg-white border border-gray-200 rounded-lg p-5 hover:border-black transition-all duration-150 cursor-pointer"
      style={{ animationDelay: `${delay}ms` }}
    >
      <div className="flex items-start justify-between mb-4">
        <div className="p-2 border border-gray-200 rounded group-hover:border-black group-hover:bg-black group-hover:text-white transition-all duration-150">
          <Icon size={18} className="text-black group-hover:text-white transition-colors" />
        </div>
        <ArrowRight size={14} className="text-gray-300 group-hover:text-black group-hover:translate-x-0.5 transition-all" />
      </div>
      <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">{label}</p>
      <AnimatedNumber value={value} className="text-2xl font-bold tracking-tight text-black mt-1 block" />
      {change !== undefined && (
        <p className={`text-xs font-medium mt-2 ${change >= 0 ? 'text-green-600' : 'text-red-600'}`}>
          {change >= 0 ? '+' : ''}{change}% from last month
        </p>
      )}
    </div>
  );

  if (href) {
    return <Link to={href}>{content}</Link>;
  }
  return content;
}

export function AnimatedNumber({ value, className = '' }) {
  const [displayValue, setDisplayValue] = useState(0);
  const [hasAnimated, setHasAnimated] = useState(false);

  useEffect(() => {
    if (hasAnimated) return;
    setHasAnimated(true);
    
    const numericValue = typeof value === 'number' ? value : parseFloat(String(value).replace(/[^0-9.]/g, '')) || 0;
    const duration = 800;
    const steps = 30;
    const stepValue = numericValue / steps;
    const stepDuration = duration / steps;

    let current = 0;
    const timer = setInterval(() => {
      current += stepValue;
      if (current >= numericValue) {
        setDisplayValue(numericValue);
        clearInterval(timer);
      } else {
        setDisplayValue(Math.floor(current));
      }
    }, stepDuration);

    return () => clearInterval(timer);
  }, [value, hasAnimated]);

  const formatValue = () => {
    if (typeof value === 'string' && value.includes('KES')) {
      return `KES ${displayValue.toLocaleString()}`;
    }
    if (typeof value === 'string' && value.includes('%')) {
      return `${displayValue}%`;
    }
    return displayValue.toLocaleString();
  };

  return <span className={className}>{formatValue()}</span>;
}

export function Card({ children, className = '', hover = false }) {
  return (
    <div className={`card ${hover ? 'card-interactive' : ''} ${className}`}>
      {children}
    </div>
  );
}

export function CardHeader({ children, className = '' }) {
  return <div className={`card-header ${className}`}>{children}</div>;
}

export function CardBody({ children, className = '' }) {
  return <div className={`card-body ${className}`}>{children}</div>;
}

export function CardFooter({ children, className = '' }) {
  return <div className={`card-footer ${className}`}>{children}</div>;
}

export function Badge({ children, variant = 'neutral', pulse = false }) {
  const variants = {
    success: 'badge-success',
    warning: 'badge-warning',
    error: 'badge-error',
    info: 'badge-info',
    neutral: 'badge-neutral',
  };
  return (
    <span className={`badge ${variants[variant]} ${pulse ? 'badge-pulse' : ''}`}>
      {children}
    </span>
  );
}

export function Button({ children, variant = 'primary', size = 'md', className = '', loading = false, ...props }) {
  const variants = {
    primary: 'btn-primary',
    secondary: 'btn-secondary',
    accent: 'btn-accent',
    ghost: 'btn-ghost',
    danger: 'btn-danger',
    success: 'btn-success',
  };
  const sizes = {
    sm: 'btn-sm',
    md: '',
    lg: 'btn-lg',
  };
  return (
    <button 
      className={`btn ${variants[variant]} ${sizes[size]} ${className}`} 
      disabled={loading || props.disabled}
      {...props}
    >
      {loading ? (
        <>
          <span className="spinner spinner-sm" />
          <span>Loading...</span>
        </>
      ) : (
        children
      )}
    </button>
  );
}

export function IconButton({ icon: Icon, variant = 'ghost', size = 'md', ...props }) {
  const variants = {
    ghost: 'btn-ghost',
    primary: 'btn-primary',
    secondary: 'btn-secondary',
    danger: 'btn-danger',
  };
  const sizes = {
    sm: 'btn-sm',
    md: '',
    lg: 'btn-lg',
  };
  return (
    <button className={`btn ${variants[variant]} ${sizes[size]} btn-icon`} {...props}>
      <Icon size={16} />
    </button>
  );
}

export function Input({ label, error, helper, className = '', icon: Icon, ...props }) {
  return (
    <div className="form-group">
      {label && <label className="label">{label}</label>}
      {Icon ? (
        <div className="input-wrapper">
          <input className={`input ${error ? 'input-error' : ''} ${className}`} {...props} />
          <Icon size={16} />
        </div>
      ) : (
        <input className={`input ${error ? 'input-error' : ''} ${className}`} {...props} />
      )}
      {error && <p className="error-text">{error}</p>}
      {helper && !error && <p className="helper-text">{helper}</p>}
    </div>
  );
}

export function Select({ label, error, options, className = '', ...props }) {
  return (
    <div className="form-group">
      {label && <label className="label">{label}</label>}
      <select className={`input select ${error ? 'input-error' : ''} ${className}`} {...props}>
        {options.map((opt) => (
          <option key={opt.value} value={opt.value}>
            {opt.label}
          </option>
        ))}
      </select>
      {error && <p className="error-text">{error}</p>}
    </div>
  );
}

export function EmptyState({ icon: Icon, title, description, action }) {
  return (
    <div className="empty-state animate-fade-in">
      {Icon && <Icon size={48} className="empty-state-icon" />}
      <h3 className="empty-state-title">{title}</h3>
      <p className="empty-state-description">{description}</p>
      {action}
    </div>
  );
}

export function SectionHeader({ title, action }) {
  return (
    <div className="section-header">
      <h3 className="section-title">{title}</h3>
      {action}
    </div>
  );
}

export function InfoRow({ label, value, className = '' }) {
  return (
    <div className={`info-row ${className}`}>
      <span className="info-label">{label}</span>
      <span className="info-value">{value}</span>
    </div>
  );
}

export function Tabs({ tabs, activeTab, onChange, count }) {
  return (
    <div className="tabs">
      {tabs.map((tab) => (
        <button
          key={tab.value}
          onClick={() => onChange(tab.value)}
          className={`tab ${activeTab === tab.value ? 'active' : ''}`}
        >
          {tab.label}
          {count && count[tab.value] > 0 && (
            <span className="tab-badge">{count[tab.value]}</span>
          )}
        </button>
      ))}
    </div>
  );
}

export function Progress({ value, max = 100, label, showPercentage = true }) {
  const percentage = (value / max) * 100;
  return (
    <div className="w-full">
      {label && (
        <div className="flex justify-between text-xs text-gray-500 mb-1">
          <span>{label}</span>
          {showPercentage && <span>{percentage.toFixed(0)}%</span>}
        </div>
      )}
      <div className="progress">
        <div className="progress-bar" style={{ width: `${percentage}%` }} />
      </div>
    </div>
  );
}

export function CircularProgress({ value, max = 100, size = 80, strokeWidth = 8 }) {
  const percentage = (value / max) * 100;
  const radius = (size - strokeWidth) / 2;
  const circumference = radius * 2 * Math.PI;
  const offset = circumference - (percentage / 100) * circumference;

  return (
    <div className="circular-progress" style={{ width: size, height: size }}>
      <svg width={size} height={size}>
        <circle
          className="circular-progress-bg"
          cx={size / 2}
          cy={size / 2}
          r={radius}
        />
        <circle
          className="circular-progress-bar"
          cx={size / 2}
          cy={size / 2}
          r={radius}
          style={{
            strokeDasharray: circumference,
            strokeDashoffset: offset,
          }}
        />
      </svg>
      <div className="circular-progress-text">{percentage.toFixed(0)}%</div>
    </div>
  );
}

export function Skeleton({ width, height, variant = 'rect' }) {
  const style = {
    width: width || '100%',
    height: height || '14px',
    borderRadius: variant === 'text' ? '4px' : variant === 'circle' ? '50%' : '6px',
  };
  return <div className="skeleton" style={style} />;
}

export function SkeletonCard() {
  return (
    <div className="skeleton-card">
      <div className="flex items-center gap-3 mb-4">
        <Skeleton width={40} height={40} variant="circle" />
        <div className="flex-1">
          <Skeleton width="60%" height={14} />
          <Skeleton width="40%" height={12} className="mt-2" />
        </div>
      </div>
      <Skeleton width="100%" height={12} />
      <Skeleton width="80%" height={12} className="mt-2" />
    </div>
  );
}

export function LoadingOverlay() {
  return (
    <div className="loading-overlay">
      <div className="spinner spinner-lg" />
    </div>
  );
}

export function PageLoading() {
  return (
    <div className="page-loading">
      <div className="spinner spinner-lg" />
      <p className="text-sm text-gray-500">Loading...</p>
    </div>
  );
}

export function Modal({ isOpen, onClose, title, children, footer, size = 'md' }) {
  if (!isOpen) return null;

  const sizes = {
    sm: 'modal-sm',
    md: '',
    lg: 'modal-lg',
    xl: 'modal-xl',
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className={`modal ${sizes[size]}`} onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2 className="modal-title">{title}</h2>
          <button className="modal-close" onClick={onClose}>
            <X size={18} />
          </button>
        </div>
        <div className="modal-body">{children}</div>
        {footer && <div className="modal-footer">{footer}</div>}
      </div>
    </div>
  );
}

export function SuccessCheckmark() {
  return (
    <div className="check-circle">
      <svg viewBox="0 0 52 52">
        <circle cx="26" cy="26" r="25" />
        <path d="M14.1 27.2l7.1 7.2 16.7-16.8" />
      </svg>
    </div>
  );
}
