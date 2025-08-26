import React, { useState, useRef } from 'react';
import {
  Box,
  Button,
  Typography,
  LinearProgress,
  Alert,
  Paper,
  IconButton,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  ListItemSecondaryAction
} from '@mui/material';
import {
  CloudUpload as UploadIcon,
  Delete as DeleteIcon,
  Description as FileIcon,
  Image as ImageIcon,
  PictureAsPdf as PdfIcon
} from '@mui/icons-material';
import axios from 'axios';

interface FileUploadProps {
  onUploadComplete?: (fileName: string) => void;
  onUploadError?: (error: string) => void;
  invoiceId?: string;
  allowedTypes?: string[];
  maxSizeMB?: number;
  multiple?: boolean;
}

interface UploadFile {
  id: string;
  file: File;
  progress: number;
  error?: string;
  uploaded?: boolean;
  fileName?: string;
}

const FileUpload: React.FC<FileUploadProps> = ({
  onUploadComplete,
  onUploadError,
  invoiceId,
  allowedTypes = ['application/pdf', 'image/jpeg', 'image/png', 'image/jpg'],
  maxSizeMB = 10,
  multiple = false
}) => {
  const [uploadFiles, setUploadFiles] = useState<UploadFile[]>([]);
  const [isDragOver, setIsDragOver] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const getFileIcon = (fileType: string) => {
    if (fileType === 'application/pdf') {
      return <PdfIcon />;
    } else if (fileType.startsWith('image/')) {
      return <ImageIcon />;
    }
    return <FileIcon />;
  };

  const validateFile = (file: File): string | null => {
    if (!allowedTypes.includes(file.type)) {
      return `File type ${file.type} is not allowed. Allowed types: ${allowedTypes.join(', ')}`;
    }
    
    if (file.size > maxSizeMB * 1024 * 1024) {
      return `File size exceeds ${maxSizeMB}MB limit`;
    }
    
    return null;
  };

  const handleFileSelect = (files: FileList | null) => {
    if (!files) return;

    const newFiles: UploadFile[] = [];
    
    for (let i = 0; i < files.length; i++) {
      const file = files[i];
      const validation = validateFile(file);
      
      const uploadFile: UploadFile = {
        id: `${Date.now()}-${i}`,
        file,
        progress: 0,
        error: validation || undefined
      };
      
      newFiles.push(uploadFile);
    }

    if (multiple) {
      setUploadFiles(prev => [...prev, ...newFiles]);
    } else {
      setUploadFiles(newFiles);
    }

    // Start upload for valid files
    newFiles.forEach(uploadFileItem => {
      if (!uploadFileItem.error) {
        uploadFile(uploadFileItem);
      }
    });
  };

  const uploadFile = async (uploadFile: UploadFile) => {
    if (!invoiceId) {
      const error = 'Invoice ID is required for file upload';
      setUploadFiles(prev => 
        prev.map(f => f.id === uploadFile.id ? { ...f, error } : f)
      );
      onUploadError?.(error);
      return;
    }

    const formData = new FormData();
    formData.append('document', uploadFile.file);

    try {
      const response = await axios.post(`/invoices/${invoiceId}/upload`, formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
        onUploadProgress: (progressEvent) => {
          if (progressEvent.total) {
            const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total);
            setUploadFiles(prev => 
              prev.map(f => f.id === uploadFile.id ? { ...f, progress } : f)
            );
          }
        },
      });

      // Mark as uploaded
      setUploadFiles(prev => 
        prev.map(f => 
          f.id === uploadFile.id 
            ? { ...f, uploaded: true, progress: 100, fileName: response.data.document_url }
            : f
        )
      );

      onUploadComplete?.(response.data.document_url);
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || 'Upload failed';
      setUploadFiles(prev => 
        prev.map(f => f.id === uploadFile.id ? { ...f, error: errorMessage } : f)
      );
      onUploadError?.(errorMessage);
    }
  };

  const handleRemoveFile = (fileId: string) => {
    setUploadFiles(prev => prev.filter(f => f.id !== fileId));
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragOver(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragOver(false);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragOver(false);
    handleFileSelect(e.dataTransfer.files);
  };

  const handleButtonClick = () => {
    fileInputRef.current?.click();
  };

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  return (
    <Box>
      <input
        type="file"
        ref={fileInputRef}
        onChange={(e) => handleFileSelect(e.target.files)}
        style={{ display: 'none' }}
        multiple={multiple}
        accept={allowedTypes.join(',')}
      />
      
      <Paper
        variant="outlined"
        sx={{
          p: 3,
          textAlign: 'center',
          border: isDragOver ? '2px dashed #1976d2' : '2px dashed #ccc',
          backgroundColor: isDragOver ? 'rgba(25, 118, 210, 0.04)' : 'transparent',
          cursor: 'pointer',
          transition: 'all 0.3s ease'
        }}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        onClick={handleButtonClick}
      >
        <UploadIcon sx={{ fontSize: 48, color: 'text.secondary', mb: 2 }} />
        <Typography variant="h6" gutterBottom>
          Drop files here or click to browse
        </Typography>
        <Typography variant="body2" color="textSecondary" gutterBottom>
          Supported formats: PDF, JPEG, PNG (max {maxSizeMB}MB)
        </Typography>
        <Button variant="contained" startIcon={<UploadIcon />} sx={{ mt: 2 }}>
          Choose Files
        </Button>
      </Paper>

      {uploadFiles.length > 0 && (
        <Box sx={{ mt: 2 }}>
          <Typography variant="h6" gutterBottom>
            Files ({uploadFiles.length})
          </Typography>
          
          <List>
            {uploadFiles.map((uploadFile) => (
              <ListItem key={uploadFile.id}>
                <ListItemIcon>
                  {getFileIcon(uploadFile.file.type)}
                </ListItemIcon>
                <ListItemText
                  primary={uploadFile.file.name}
                  secondary={
                    <Box>
                      <Typography variant="caption" display="block">
                        {formatFileSize(uploadFile.file.size)}
                      </Typography>
                      {uploadFile.error && (
                        <Alert severity="error" sx={{ mt: 1 }}>
                          {uploadFile.error}
                        </Alert>
                      )}
                      {uploadFile.uploaded && (
                        <Alert severity="success" sx={{ mt: 1 }}>
                          Upload completed successfully
                        </Alert>
                      )}
                      {!uploadFile.error && !uploadFile.uploaded && uploadFile.progress > 0 && (
                        <Box sx={{ mt: 1 }}>
                          <LinearProgress variant="determinate" value={uploadFile.progress} />
                          <Typography variant="caption" color="textSecondary">
                            {uploadFile.progress}% completed
                          </Typography>
                        </Box>
                      )}
                    </Box>
                  }
                />
                <ListItemSecondaryAction>
                  <IconButton
                    edge="end"
                    onClick={() => handleRemoveFile(uploadFile.id)}
                    disabled={uploadFile.progress > 0 && !uploadFile.uploaded && !uploadFile.error}
                  >
                    <DeleteIcon />
                  </IconButton>
                </ListItemSecondaryAction>
              </ListItem>
            ))}
          </List>
        </Box>
      )}
    </Box>
  );
};

export default FileUpload;
