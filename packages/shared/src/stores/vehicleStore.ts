import { create } from 'zustand';
import type { Vehicle } from '../types';
import { vehicleApi } from '../api';

interface VehicleState {
  vehicles: Vehicle[];
  currentVehicle: Vehicle | null;
  isLoading: boolean;

  fetchVehicles: () => Promise<void>;
  setCurrentVehicle: (vehicle: Vehicle | null) => void;
  addVehicle: (vehicle: Vehicle) => void;
  updateVehicle: (vehicle: Vehicle) => void;
  removeVehicle: (id: string) => void;
}

export const useVehicleStore = create<VehicleState>((set, get) => ({
  vehicles: [],
  currentVehicle: null,
  isLoading: false,

  fetchVehicles: async () => {
    set({ isLoading: true });
    try {
      const { data } = await vehicleApi.list();
      const vehicles = data.data;
      set({ vehicles });

      // 自动选择默认车辆
      if (!get().currentVehicle && vehicles.length > 0) {
        const defaultVehicle = vehicles.find((v) => v.is_default) || vehicles[0];
        set({ currentVehicle: defaultVehicle });
      }
    } finally {
      set({ isLoading: false });
    }
  },

  setCurrentVehicle: (vehicle) => set({ currentVehicle: vehicle }),

  addVehicle: (vehicle) =>
    set((state) => ({ vehicles: [...state.vehicles, vehicle] })),

  updateVehicle: (vehicle) =>
    set((state) => ({
      vehicles: state.vehicles.map((v) => (v.id === vehicle.id ? vehicle : v)),
      currentVehicle:
        state.currentVehicle?.id === vehicle.id ? vehicle : state.currentVehicle,
    })),

  removeVehicle: (id) =>
    set((state) => ({
      vehicles: state.vehicles.filter((v) => v.id !== id),
      currentVehicle: state.currentVehicle?.id === id ? null : state.currentVehicle,
    })),
}));
