import React from 'react';
import { View, Text, StyleSheet, FlatList } from 'react-native';
import { Button, Card } from 'react-native-paper';
import { useNavigation } from '@react-navigation/native';
import { StackNavigationProp } from '@react-navigation/stack';
import { RootStackParamList } from '@/navigation/AppNavigator';

type NavigationProp = StackNavigationProp<RootStackParamList, 'Organizations'>;

const mockOrganizations = [
  { id: 1, name: 'ООО "Ромашка"', status: 'approved' },
  { id: 2, name: 'ИП Иванов', status: 'pending' },
  { id: 3, name: 'АО "Технологии"', status: 'approved' },
];

export default function OrganizationsScreen() {
  const navigation = useNavigation<NavigationProp>();

  const renderItem = ({ item }: { item: any }) => (
    <Card style={styles.card}>
      <Card.Content>
        <Text style={styles.orgName}>{item.name}</Text>
        <Text style={styles.orgStatus}>Статус: {item.status}</Text>
      </Card.Content>
    </Card>
  );

  return (
    <View style={styles.container}>
      <Text style={styles.title}>Мои организации</Text>
      
      <FlatList
        data={mockOrganizations}
        renderItem={renderItem}
        keyExtractor={(item) => item.id.toString()}
        contentContainerStyle={styles.list}
      />
      
      <Button
        mode="contained"
        onPress={() => console.log('Create organization')}
        style={styles.createButton}
      >
        Создать организацию
      </Button>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    padding: 16,
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 20,
  },
  list: {
    paddingBottom: 16,
  },
  card: {
    marginBottom: 12,
  },
  orgName: {
    fontSize: 16,
    fontWeight: '600',
  },
  orgStatus: {
    fontSize: 14,
    color: '#666',
    marginTop: 4,
  },
  createButton: {
    marginTop: 16,
  },
});
