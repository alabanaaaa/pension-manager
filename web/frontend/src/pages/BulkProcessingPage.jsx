import { useState } from 'react';
import { bulk } from '../lib/api';
import { Loader2, Upload, FileSpreadsheet, CheckCircle, AlertCircle, Calendar, RefreshCw } from 'lucide-react';

export default function BulkProcessingPage() {
  const [file, setFile] = useState(null);
  const [uploading, setUploading] = useState(false);
  const [result, setResult] = useState(null);
  const [processing, setProcessing] = useState(false);
  const [processResult, setProcessResult] = useState(null);
  const [year, setYear] = useState(new Date().getFullYear());

  const handleImport = async () => {
    if (!file) return;
    setUploading(true);
    setResult(null);
    try {
      const formData = new FormData();
      formData.append('file', file);
      const res = await bulk.importMembers(formData);
      setResult(res.data);
    } catch (err) {
      setResult({ error: err.response?.data?.error || 'Import failed' });
    }
    finally { setUploading(false); }
  };

  const handleAnnualPosting = async () => {
    setProcessing(true);
    setProcessResult(null);
    try {
      const res = await bulk.annualPosting(year);
      setProcessResult(res.data);
    } catch (err) {
      setProcessResult({ error: err.response?.data?.error || 'Processing failed' });
    }
    finally { setProcessing(false); }
  };

  return (
    <div className="space-y-8 animate-fade-in-up">
      <div>
        <h1 className="text-3xl font-semibold tracking-tight text-neutral-900">Bulk Processing</h1>
        <p className="text-neutral-500 mt-2 text-base">Import members, batch statements, and annual posting</p>
      </div>

      {/* CSV Import */}
      <div className="bg-white rounded-2xl border border-neutral-50 overflow-hidden">
        <div className="flex items-center justify-between px-6 py-5 border-b border-neutral-50">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-blue-50"><Upload size={20} className="text-blue-600" /></div>
            <h2 className="text-lg font-semibold tracking-tight text-neutral-900">Import Members (CSV)</h2>
          </div>
        </div>
        <div className="p-6 space-y-6">
          <div className="border-2 border-dashed border-neutral-200 rounded-xl p-8 text-center hover:border-neutral-300 transition-colors">
            <FileSpreadsheet size={32} className="mx-auto text-neutral-300 mb-3" />
            <p className="text-sm text-neutral-500 mb-2">
              {file ? file.name : 'Drag & drop your CSV file here, or click to browse'}
            </p>
            <input
              type="file"
              accept=".csv"
              onChange={e => setFile(e.target.files[0])}
              className="hidden"
              id="csv-upload"
            />
            <label htmlFor="csv-upload" className="inline-flex items-center gap-2 px-4 py-2.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm font-medium hover:bg-neutral-100 cursor-pointer transition-all">
              Choose File
            </label>
          </div>
          <button
            onClick={handleImport}
            disabled={uploading || !file}
            className="btn-hover flex items-center gap-2 px-6 py-3.5 bg-neutral-900 text-white rounded-xl text-sm font-medium hover:bg-neutral-800 disabled:opacity-50 transition-all"
          >
            {uploading ? <Loader2 size={16} className="animate-spin" /> : <Upload size={16} />}
            Import Members
          </button>

          {result && (
            <div className={`p-5 rounded-xl ${result.error ? 'bg-red-50 text-red-700' : 'bg-emerald-50 text-emerald-700'}`}>
              <div className="flex items-center gap-3">
                {result.error ? <AlertCircle size={18} /> : <CheckCircle size={18} />}
                <div>
                  <p className="font-medium">{result.error ? 'Import Failed' : 'Import Complete'}</p>
                  {!result.error && (
                    <p className="text-sm mt-1">
                      {result.success} succeeded · {result.failed} failed out of {result.total_rows} rows
                    </p>
                  )}
                  {result.errors && result.errors.length > 0 && (
                    <div className="mt-3 text-xs space-y-1">
                      {result.errors.slice(0, 5).map((e, i) => (
                        <p key={i}>Row {e.row}: {e.field} — {e.reason}</p>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Annual Posting */}
      <div className="bg-white rounded-2xl border border-neutral-50 overflow-hidden">
        <div className="flex items-center justify-between px-6 py-5 border-b border-neutral-50">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-emerald-50"><Calendar size={20} className="text-emerald-600" /></div>
            <h2 className="text-lg font-semibold tracking-tight text-neutral-900">Annual Posting</h2>
          </div>
        </div>
        <div className="p-6 space-y-6">
          <div className="flex items-center gap-4">
            <input
              type="number"
              value={year}
              onChange={e => setYear(parseInt(e.target.value))}
              className="w-32 px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all"
            />
            <button
              onClick={handleAnnualPosting}
              disabled={processing}
              className="btn-hover flex items-center gap-2 px-6 py-3.5 bg-neutral-900 text-white rounded-xl text-sm font-medium hover:bg-neutral-800 disabled:opacity-50 transition-all"
            >
              {processing ? <Loader2 size={16} className="animate-spin" /> : <RefreshCw size={16} />}
              Post Annual Contributions
            </button>
          </div>

          {processResult && (
            <div className={`p-5 rounded-xl ${processResult.error ? 'bg-red-50 text-red-700' : 'bg-emerald-50 text-emerald-700'}`}>
              <div className="flex items-center gap-3">
                {processResult.error ? <AlertCircle size={18} /> : <CheckCircle size={18} />}
                <div>
                  <p className="font-medium">{processResult.error ? 'Processing Failed' : 'Annual Posting Complete'}</p>
                  {!processResult.error && (
                    <p className="text-sm mt-1">{processResult.processed} members processed</p>
                  )}
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
