import React, { useState } from 'react';
import {
  AppBar,
  Toolbar,
  Typography,
  Button,
  Box,
  Menu,
  MenuItem,
  Avatar,
  IconButton,
  Chip,
} from '@mui/material';
import {
  AccountBalanceWallet,
  Person,
  ExitToApp,
  Dashboard,
  BusinessCenter,
  Assessment,
} from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';
import { useWeb3 } from '../../contexts/Web3Context';

const Navbar: React.FC = () => {
  const { user, logout } = useAuth();
  const { account, isConnected, connectWallet, disconnectWallet } = useWeb3();
  const navigate = useNavigate();
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);

  const handleMenuClick = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleMenuClose = () => {
    setAnchorEl(null);
  };

  const handleLogout = () => {
    logout();
    disconnectWallet();
    handleMenuClose();
    navigate('/');
  };

  const handleWalletConnect = async () => {
    try {
      await connectWallet();
    } catch (error: any) {
      console.error('Failed to connect wallet:', error);
    }
  };

  const getDashboardPath = () => {
    if (!user) return '/';
    switch (user.role) {
      case 'sme':
        return '/sme-dashboard';
      case 'investor':
        return '/investor-dashboard';
      case 'admin':
        return '/admin-dashboard';
      default:
        return '/';
    }
  };

  const formatAddress = (address: string) => {
    return `${address.slice(0, 6)}...${address.slice(-4)}`;
  };

  return (
    <AppBar position="static" elevation={1}>
      <Toolbar>
        <Typography
          variant="h6"
          component="div"
          sx={{ flexGrow: 1, cursor: 'pointer' }}
          onClick={() => navigate('/')}
        >
          Invoice Finance Platform
        </Typography>

        {user ? (
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
            {/* Dashboard Button */}
            <Button
              color="inherit"
              startIcon={<Dashboard />}
              onClick={() => navigate(getDashboardPath())}
            >
              Dashboard
            </Button>

            {/* Role-specific navigation */}
            {user.role === 'sme' && (
              <Button
                color="inherit"
                startIcon={<BusinessCenter />}
                onClick={() => navigate('/invoices')}
              >
                Invoices
              </Button>
            )}

            {user.role === 'investor' && (
              <>
                <Button
                  color="inherit"
                  startIcon={<Assessment />}
                  onClick={() => navigate('/marketplace')}
                >
                  Marketplace
                </Button>
                <Button
                  color="inherit"
                  startIcon={<BusinessCenter />}
                  onClick={() => navigate('/investments')}
                >
                  Investments
                </Button>
              </>
            )}

            {/* Wallet Connection */}
            {isConnected ? (
              <Chip
                icon={<AccountBalanceWallet />}
                label={formatAddress(account!)}
                variant="outlined"
                sx={{ color: 'white', borderColor: 'white' }}
                onClick={disconnectWallet}
              />
            ) : (
              <Button
                variant="outlined"
                color="inherit"
                startIcon={<AccountBalanceWallet />}
                onClick={handleWalletConnect}
                sx={{ borderColor: 'white' }}
              >
                Connect Wallet
              </Button>
            )}

            {/* User Menu */}
            <IconButton color="inherit" onClick={handleMenuClick}>
              <Avatar sx={{ width: 32, height: 32, bgcolor: 'secondary.main' }}>
                {user.first_name.charAt(0)}
              </Avatar>
            </IconButton>

            <Menu
              anchorEl={anchorEl}
              open={Boolean(anchorEl)}
              onClose={handleMenuClose}
            >
              <MenuItem onClick={() => { navigate('/profile'); handleMenuClose(); }}>
                <Person sx={{ mr: 1 }} />
                Profile
              </MenuItem>
              <MenuItem onClick={handleLogout}>
                <ExitToApp sx={{ mr: 1 }} />
                Logout
              </MenuItem>
            </Menu>
          </Box>
        ) : (
          <Box sx={{ display: 'flex', gap: 1 }}>
            <Button color="inherit" onClick={() => navigate('/login')}>
              Login
            </Button>
            <Button
              variant="outlined"
              color="inherit"
              onClick={() => navigate('/register')}
              sx={{ borderColor: 'white' }}
            >
              Register
            </Button>
          </Box>
        )}
      </Toolbar>
    </AppBar>
  );
};

export default Navbar;
