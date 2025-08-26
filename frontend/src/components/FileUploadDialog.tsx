import React, { useState } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Typography,
  Box,
  Alert
} from '@mui/material';
import FileUpload from './FileUpload';

interface FileUploadDialogProps {
  open: boolean;
  onClose: () => void;
  invoiceId: string;
  invoiceNumber: string;
  onUploadSuccess?: () => void;
}

const FileUploadDialog: React.FC<FileUploadDialogProps> = ({
  open,
  onClose,
  invoiceId,
  invoiceNumber,
  onUploadSuccess
}) => {
  const [uploadSuccess, setUploadSuccess] = useState(false);
  const [error, setError] = useState('');

  const handleUploadComplete = (fileName: string) => {
    setUploadSuccess(true);
    setError('');
    onUploadSuccess?.();
  };

  const handleUploadError = (error: string) => {
    setError(error);
    setUploadSuccess(false);
  };

  const handleClose = () => {
    setUploadSuccess(false);
    setError('');
    onClose();
  };

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="md" fullWidth>
      <DialogTitle>
        Upload Document for Invoice {invoiceNumber}
      </DialogTitle>
      <DialogContent>
        <Typography variant="body2" color="textSecondary" sx={{ mb: 3 }}>
          Upload the invoice document to verify and process your invoice. 
          Supported formats: PDF, JPEG, PNG (max 10MB)
        </Typography>
        
        {error && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {error}
          </Alert>
        )}
        
        {uploadSuccess && (
          <Alert severity="success" sx={{ mb: 2 }}>
            Document uploaded successfully! Your invoice will be processed for verification.
          </Alert>
        )}
        
        <FileUpload
          invoiceId={invoiceId}
          onUploadComplete={handleUploadComplete}
          onUploadError={handleUploadError}
          multiple={false}
        />
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose}>
          {uploadSuccess ? 'Close' : 'Cancel'}
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export default FileUploadDialog;
