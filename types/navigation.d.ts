// src/types/navigation.d.ts
import { RootStackParamList } from '@/navigation/AppNavigator';

declare global {
  namespace ReactNavigation {
    interface RootParamList extends RootStackParamList {}
  }
}