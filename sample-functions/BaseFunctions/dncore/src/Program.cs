using System;
using System.Text;

namespace root
{
    class Program
    {
        private static string getStdin() {
            StringBuilder buffer = new StringBuilder();
            string s;
            while ((s = Console.ReadLine()) != null)
            {
                buffer.AppendLine(s);
            }
            return buffer.ToString();
        }

        static void Main(string[] args)
        {
            string buffer = getStdin();

            Console.WriteLine(buffer);
        }
    }
}
