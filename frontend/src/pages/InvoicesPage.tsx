import React, { useState, useEffect } from 'react';
import {
  Container,
  Typography,
  Box,
  Button,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Chip,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Grid,
  Alert,
  CircularProgress,
  Card,
  CardContent,
  Fab,
  Menu,
  MenuItem
} from '@mui/material';
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  Visibility as ViewIcon,
  Upload as UploadIcon,
  MoreVert as MoreVertIcon,
  Description as DocumentIcon
} from '@mui/icons-material';
import { useAuth } from '../contexts/AuthContext';
import FileUploadDialog from '../components/FileUploadDialog';
import axios from 'axios';

interface Invoice {
  id: string;
  invoice_number: string;
  customer_name: string;
  customer_email?: string;
  invoice_amount: number;
  due_date: string;
  issue_date: string;
  description?: string;
  status: 'pending' | 'verified' | 'financed' | 'paid' | 'overdue' | 'rejected';
  document_url?: string;
  created_at: string;
}

const InvoicesPage: React.FC = () => {
  const [invoices, setInvoices] = useState<Invoice[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [openDialog, setOpenDialog] = useState(false);
  const [selectedInvoice, setSelectedInvoice] = useState<Invoice | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [selectedInvoiceForMenu, setSelectedInvoiceForMenu] = useState<Invoice | null>(null);
  const [uploadDialogOpen, setUploadDialogOpen] = useState(false);
  const [selectedInvoiceForUpload, setSelectedInvoiceForUpload] = useState<Invoice | null>(null);
  const { user } = useAuth();

  const [formData, setFormData] = useState({
    invoice_number: '',
    customer_name: '',
    customer_email: '',
    invoice_amount: '',
    due_date: '',
    issue_date: '',
    description: ''
  });

  useEffect(() => {
    fetchInvoices();
  }, []);

  const fetchInvoices = async () => {
    try {
      setLoading(true);
      const response = await axios.get('/invoices');
      setInvoices(response.data);
    } catch (err: any) {
      setError('Failed to fetch invoices');
      console.error('Error fetching invoices:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleOpenDialog = (invoice?: Invoice) => {
    if (invoice) {
      setSelectedInvoice(invoice);
      setIsEditing(true);
      setFormData({
        invoice_number: invoice.invoice_number,
        customer_name: invoice.customer_name,
        customer_email: invoice.customer_email || '',
        invoice_amount: invoice.invoice_amount.toString(),
        due_date: invoice.due_date.split('T')[0],
        issue_date: invoice.issue_date.split('T')[0],
        description: invoice.description || ''
      });
    } else {
      setSelectedInvoice(null);
      setIsEditing(false);
      setFormData({
        invoice_number: '',
        customer_name: '',
        customer_email: '',
        invoice_amount: '',
        due_date: new Date().toISOString().split('T')[0],
        issue_date: new Date().toISOString().split('T')[0],
        description: ''
      });
    }
    setOpenDialog(true);
  };

  const handleCloseDialog = () => {
    setOpenDialog(false);
    setSelectedInvoice(null);
    setIsEditing(false);
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async () => {
    try {
      const invoiceData = {
        ...formData,
        invoice_amount: parseFloat(formData.invoice_amount),
        due_date: new Date(formData.due_date).toISOString(),
        issue_date: new Date(formData.issue_date).toISOString()
      };

      if (isEditing && selectedInvoice) {
        await axios.put(`/invoices/${selectedInvoice.id}`, invoiceData);
      } else {
        await axios.post('/invoices', invoiceData);
      }
      
      fetchInvoices();
      handleCloseDialog();
    } catch (err: any) {
      setError(`Failed to ${isEditing ? 'update' : 'create'} invoice`);
      console.error('Error saving invoice:', err);
    }
  };

  const handleDelete = async (invoice: Invoice) => {
    if (window.confirm('Are you sure you want to delete this invoice?')) {
      try {
        await axios.delete(`/invoices/${invoice.id}`);
        fetchInvoices();
      } catch (err: any) {
        setError('Failed to delete invoice');
        console.error('Error deleting invoice:', err);
      }
    }
  };

  const getStatusColor = (status: string) => {
    const statusColors: Record<string, 'default' | 'primary' | 'secondary' | 'error' | 'info' | 'success' | 'warning'> = {
      pending: 'default',
      verified: 'info',
      financed: 'primary',
      paid: 'success',
      overdue: 'error',
      rejected: 'error'
    };
    return statusColors[status] || 'default';
  };

  const handleMenuClick = (event: React.MouseEvent<HTMLElement>, invoice: Invoice) => {
    setAnchorEl(event.currentTarget);
    setSelectedInvoiceForMenu(invoice);
  };

  const handleMenuClose = () => {
    setAnchorEl(null);
    setSelectedInvoiceForMenu(null);
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(amount);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString();
  };

  if (loading) {
    return (
      <Container maxWidth="lg">
        <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '50vh' }}>
          <CircularProgress />
        </Box>
      </Container>
    );
  }

  return (
    <Container maxWidth="lg">
      <Box sx={{ mt: 4, mb: 4 }}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 4 }}>
          <Typography variant="h4" component="h1">
            My Invoices
          </Typography>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => handleOpenDialog()}
          >
            Create Invoice
          </Button>
        </Box>

        {error && (
          <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError('')}>
            {error}
          </Alert>
        )}

        {/* Statistics Cards */}
        <Grid container spacing={3} sx={{ mb: 4 }}>
          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Typography color="textSecondary" gutterBottom>
                  Total Invoices
                </Typography>
                <Typography variant="h4">
                  {invoices.length}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Typography color="textSecondary" gutterBottom>
                  Total Value
                </Typography>
                <Typography variant="h4">
                  {formatCurrency(invoices.reduce((sum, inv) => sum + inv.invoice_amount, 0))}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Typography color="textSecondary" gutterBottom>
                  Pending
                </Typography>
                <Typography variant="h4">
                  {invoices.filter(inv => inv.status === 'pending').length}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Typography color="textSecondary" gutterBottom>
                  Financed
                </Typography>
                <Typography variant="h4">
                  {invoices.filter(inv => inv.status === 'financed').length}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
        </Grid>

        {/* Invoices Table */}
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Invoice #</TableCell>
                <TableCell>Customer</TableCell>
                <TableCell>Amount</TableCell>
                <TableCell>Issue Date</TableCell>
                <TableCell>Due Date</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>Document</TableCell>
                <TableCell>Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {invoices.map((invoice) => (
                <TableRow key={invoice.id}>
                  <TableCell>{invoice.invoice_number}</TableCell>
                  <TableCell>
                    <Box>
                      <Typography variant="body2">{invoice.customer_name}</Typography>
                      {invoice.customer_email && (
                        <Typography variant="caption" color="textSecondary">
                          {invoice.customer_email}
                        </Typography>
                      )}
                    </Box>
                  </TableCell>
                  <TableCell>{formatCurrency(invoice.invoice_amount)}</TableCell>
                  <TableCell>{formatDate(invoice.issue_date)}</TableCell>
                  <TableCell>{formatDate(invoice.due_date)}</TableCell>
                  <TableCell>
                    <Chip
                      label={invoice.status.charAt(0).toUpperCase() + invoice.status.slice(1)}
                      color={getStatusColor(invoice.status)}
                      size="small"
                    />
                  </TableCell>
                  <TableCell>
                    {invoice.document_url ? (
                      <IconButton size="small" color="primary">
                        <DocumentIcon />
                      </IconButton>
                    ) : (
                      <Typography variant="caption" color="textSecondary">
                        No document
                      </Typography>
                    )}
                  </TableCell>
                  <TableCell>
                    <IconButton
                      size="small"
                      onClick={(e) => handleMenuClick(e, invoice)}
                    >
                      <MoreVertIcon />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>

        {invoices.length === 0 && (
          <Box sx={{ textAlign: 'center', py: 8 }}>
            <Typography variant="h6" color="textSecondary">
              No invoices found
            </Typography>
            <Typography variant="body2" color="textSecondary" sx={{ mb: 2 }}>
              Create your first invoice to get started
            </Typography>
            <Button
              variant="contained"
              startIcon={<AddIcon />}
              onClick={() => handleOpenDialog()}
            >
              Create Invoice
            </Button>
          </Box>
        )}
      </Box>

      {/* Invoice Form Dialog */}
      <Dialog open={openDialog} onClose={handleCloseDialog} maxWidth="md" fullWidth>
        <DialogTitle>
          {isEditing ? 'Edit Invoice' : 'Create New Invoice'}
        </DialogTitle>
        <DialogContent>
          <Grid container spacing={3} sx={{ mt: 1 }}>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="Invoice Number"
                name="invoice_number"
                value={formData.invoice_number}
                onChange={handleInputChange}
                required
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="Amount"
                name="invoice_amount"
                type="number"
                value={formData.invoice_amount}
                onChange={handleInputChange}
                required
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="Customer Name"
                name="customer_name"
                value={formData.customer_name}
                onChange={handleInputChange}
                required
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="Customer Email"
                name="customer_email"
                type="email"
                value={formData.customer_email}
                onChange={handleInputChange}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="Issue Date"
                name="issue_date"
                type="date"
                value={formData.issue_date}
                onChange={handleInputChange}
                InputLabelProps={{ shrink: true }}
                required
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="Due Date"
                name="due_date"
                type="date"
                value={formData.due_date}
                onChange={handleInputChange}
                InputLabelProps={{ shrink: true }}
                required
              />
            </Grid>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="Description"
                name="description"
                multiline
                rows={3}
                value={formData.description}
                onChange={handleInputChange}
              />
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseDialog}>Cancel</Button>
          <Button onClick={handleSubmit} variant="contained">
            {isEditing ? 'Update' : 'Create'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Action Menu */}
      <Menu
        anchorEl={anchorEl}
        open={Boolean(anchorEl)}
        onClose={handleMenuClose}
      >
        <MenuItem
          onClick={() => {
            if (selectedInvoiceForMenu) handleOpenDialog(selectedInvoiceForMenu);
            handleMenuClose();
          }}
        >
          <EditIcon fontSize="small" sx={{ mr: 1 }} />
          Edit
        </MenuItem>
        <MenuItem
          onClick={() => {
            if (selectedInvoiceForMenu) {
              setSelectedInvoiceForUpload(selectedInvoiceForMenu);
              setUploadDialogOpen(true);
            }
            handleMenuClose();
          }}
        >
          <UploadIcon fontSize="small" sx={{ mr: 1 }} />
          Upload Document
        </MenuItem>
        <MenuItem
          onClick={() => {
            if (selectedInvoiceForMenu) handleDelete(selectedInvoiceForMenu);
            handleMenuClose();
          }}
          sx={{ color: 'error.main' }}
        >
          <DeleteIcon fontSize="small" sx={{ mr: 1 }} />
          Delete
        </MenuItem>
      </Menu>

      {/* File Upload Dialog */}
      {selectedInvoiceForUpload && (
        <FileUploadDialog
          open={uploadDialogOpen}
          onClose={() => {
            setUploadDialogOpen(false);
            setSelectedInvoiceForUpload(null);
          }}
          invoiceId={selectedInvoiceForUpload.id}
          invoiceNumber={selectedInvoiceForUpload.invoice_number}
          onUploadSuccess={fetchInvoices}
        />
      )}
    </Container>
  );
};

export default InvoicesPage;
