import { create } from 'zustand';
import type { Vehicle } from '../types';
import { vehicleApi } from '../api';

interface VehicleState {
  vehicles: Vehicle[];
  isLoading: boolean;

  fetchVehicles: () => Promise<void>;
  addVehicle: (vehicle: Vehicle) => void;
  updateVehicle: (vehicle: Vehicle) => void;
  removeVehicle: (id: string) => void;
  reset: () => void;
}

export const useVehicleStore = create<VehicleState>((set) => ({
  vehicles: [],
  isLoading: false,

  fetchVehicles: async () => {
    set({ isLoading: true });
    try {
      const { data } = await vehicleApi.list();
      set({ vehicles: data.data });
    } finally {
      set({ isLoading: false });
    }
  },

  addVehicle: (vehicle) =>
    set((state) => ({ vehicles: [...state.vehicles, vehicle] })),

  updateVehicle: (vehicle) =>
    set((state) => ({
      vehicles: state.vehicles.map((v) => (v.id === vehicle.id ? vehicle : v)),
    })),

  removeVehicle: (id) =>
    set((state) => ({
      vehicles: state.vehicles.filter((v) => v.id !== id),
    })),

  reset: () => set({ vehicles: [], isLoading: false }),
}));
