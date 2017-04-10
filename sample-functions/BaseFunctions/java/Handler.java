import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;

public class Handler {

    public static void main(String[] args) {
        try {

            String input = readStdin();
            System.out.print(input);

        } catch(IOException e) {
            e.printStackTrace();
        }
    }

    private static String readStdin() throws IOException {
        BufferedReader br = new BufferedReader(new InputStreamReader(System.in));
        String input = "";
        while(true) {
            String line = br.readLine();
            if(line==null) {
                break;
            }
            input = input + line + "\n";
        }
        return input;
    }
}
