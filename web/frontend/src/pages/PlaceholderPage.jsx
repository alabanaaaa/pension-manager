import { Construction } from 'lucide-react';

export default function PlaceholderPage() {
  return (
    <div className="min-h-[60vh] flex items-center justify-center">
      <div className="text-center max-w-md">
        <div className="w-20 h-20 border-2 border-dashed border-gray-200 rounded-2xl flex items-center justify-center mx-auto mb-6">
          <Construction size={32} className="text-gray-400" />
        </div>
        <div className="inline-block px-4 py-1.5 bg-black text-white text-xs font-semibold uppercase tracking-wider rounded mb-4">
          In Development
        </div>
        <h1 className="text-2xl font-bold tracking-tight text-black mb-2">Coming Soon</h1>
        <p className="text-sm text-gray-500">
          This feature is currently under development and will be available soon.
        </p>
      </div>
    </div>
  );
}
