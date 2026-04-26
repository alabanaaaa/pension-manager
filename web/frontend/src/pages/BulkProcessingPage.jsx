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
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-black">Bulk Processing</h1>
        <p className="text-sm text-gray-500 mt-1">Import members, batch statements, and annual posting</p>
      </div>

      {/* CSV Import */}
      <div className="card">
        <div className="card-header flex items-center gap-3">
          <Upload size={18} className="text-gray-500" />
          <h2 className="text-base font-semibold text-black">Import Members (CSV)</h2>
        </div>
        <div className="card-body space-y-4">
          <div className="border-2 border-dashed border-gray-200 rounded-lg p-8 text-center hover:border-black transition-colors">
            <FileSpreadsheet size={32} className="mx-auto text-gray-300 mb-3" />
            <p className="text-sm text-gray-500 mb-3">
              {file ? file.name : 'Drag & drop your CSV file here, or click to browse'}
            </p>
            <input
              type="file"
              accept=".csv"
              onChange={e => setFile(e.target.files[0])}
              className="hidden"
              id="csv-upload"
            />
            <label htmlFor="csv-upload" className="btn btn-secondary cursor-pointer">
              Choose File
            </label>
          </div>
          <button
            onClick={handleImport}
            disabled={uploading || !file}
            className="btn btn-primary"
          >
            {uploading ? <Loader2 size={16} className="animate-spin" /> : <Upload size={16} />}
            Import Members
          </button>

          {result && (
            <div className={`p-4 rounded-lg flex items-start gap-3 ${result.error ? 'bg-red-50 text-red-700' : 'bg-green-50 text-green-700'}`}>
              {result.error ? <AlertCircle size={18} className="flex-shrink-0 mt-0.5" /> : <CheckCircle size={18} className="flex-shrink-0 mt-0.5" />}
              <div>
                <p className="font-semibold">{result.error ? 'Import Failed' : 'Import Complete'}</p>
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
          )}
        </div>
      </div>

      {/* Annual Posting */}
      <div className="card">
        <div className="card-header flex items-center gap-3">
          <Calendar size={18} className="text-gray-500" />
          <h2 className="text-base font-semibold text-black">Annual Posting</h2>
        </div>
        <div className="card-body space-y-4">
          <div className="flex items-center gap-4">
            <input
              type="number"
              value={year}
              onChange={e => setYear(parseInt(e.target.value))}
              className="input w-32"
            />
            <button
              onClick={handleAnnualPosting}
              disabled={processing}
              className="btn btn-primary"
            >
              {processing ? <Loader2 size={16} className="animate-spin" /> : <RefreshCw size={16} />}
              Post Annual Contributions
            </button>
          </div>

          {processResult && (
            <div className={`p-4 rounded-lg flex items-start gap-3 ${processResult.error ? 'bg-red-50 text-red-700' : 'bg-green-50 text-green-700'}`}>
              {processResult.error ? <AlertCircle size={18} className="flex-shrink-0 mt-0.5" /> : <CheckCircle size={18} className="flex-shrink-0 mt-0.5" />}
              <div>
                <p className="font-semibold">{processResult.error ? 'Processing Failed' : 'Annual Posting Complete'}</p>
                {!processResult.error && (
                  <p className="text-sm mt-1">{processResult.processed} members processed</p>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
