
public class BuscaBinaria {

	public int posicao;
	
	BuscaBinaria(int[] array, int Objetivo){
		this.posicao = BinarySearch(Objetivo, array);
	}
	public int GetResultado() {
		return this.posicao;
	}
	public int BinarySearch(int objetivo, int[] array) {
		
		int comeco = 0;
		int fim = array.length - 1; 
        while (comeco <= fim) { 
            int meio = comeco + (fim-comeco)/2; 
            if (array[meio] == objetivo) 
                return meio; 
            if (array[meio] < objetivo) 
                comeco = meio + 1; 
            else
            	fim = meio - 1; 
        }
        return -1; 
    }
}


